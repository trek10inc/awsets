package resource

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/fatih/structs"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"gopkg.in/yaml.v2"
)

type Identifier struct {
	Account string
	Region  string
	Id      string
	Version string
	Type    ResourceType
}

type Resource struct {
	Identifier
	Name       string
	Attributes map[string]interface{}
	Tags       map[string]string
	Relations  []Identifier
}

type Group struct {
	sync.RWMutex
	Resources map[Identifier]Resource
}

func NewGlobal(ctx context.AWSetsCtx, kind ResourceType, id, name, rawObject interface{}) Resource {
	return makeResource(ctx.AccountId, "aws-global", kind, id, name, "", rawObject)
}

func New(ctx context.AWSetsCtx, kind ResourceType, id, name, rawObject interface{}) Resource {
	return makeResource(ctx.AccountId, ctx.Region(), kind, id, name, "", rawObject)
}

func NewVersion(ctx context.AWSetsCtx, kind ResourceType, id, name, version, rawObject interface{}) Resource {
	return makeResource(ctx.AccountId, ctx.Region(), kind, id, name, version, rawObject)
}

func makeResource(account, region string, kind ResourceType, iId, iName, iVersion, rawObject interface{}) Resource {

	id := toString(iId)
	name := toString(iName)
	version := toString(iVersion)

	var asMap map[string]interface{}
	if structs.IsStruct(rawObject) {
		asMap = structs.Map(rawObject)
	} else {
		asMap = rawObject.(map[string]interface{})
	}

	if _, ok := asMap["ResultMetadata"]; ok {
		delete(asMap, "ResultMetadata")
	}
	resource := Resource{
		Identifier: Identifier{
			Account: account,
			Region:  region,
			Id:      id,
			Version: version,
			Type:    kind,
		},
		Name:       name,
		Attributes: asMap,
		Tags:       make(map[string]string),
	}
	if strings.Contains(id, "arn:") && kind != SsmPatchBaseline {
		fmt.Fprintf(os.Stderr, "new resource contains arn: %s - %s\n", kind.String(), id)
	}
	if tags, ok := asMap["Tags"]; ok {
		switch t := tags.(type) {
		case []interface{}:
			for _, v := range t {
				tag := v.(map[string]interface{})
				key := tag["Key"].(*string)
				value := tag["Value"].(*string)
				resource.Tags[*key] = *value
			}
		case map[string]string:
			for k, v := range t {
				resource.Tags[k] = v
			}
		case map[string]*string:
			for k, v := range t {
				resource.Tags[k] = *v
			}
		case nil:
			// no op
		default:
			fmt.Fprintf(os.Stderr, "Unknown tag type: %T\n", t)
		}
	}
	return resource
}

func (r *Resource) AddAttribute(key string, value interface{}) {
	if value != nil {
		if structs.IsStruct(value) {
			asMap := structs.Map(value)
			if _, ok := asMap["ResultMetadata"]; ok {
				delete(asMap, "ResultMetadata")
			}
			r.Attributes[key] = asMap
		} else {
			r.Attributes[key] = value
		}
	}
}

func (r *Resource) AddARNRelation(kind ResourceType, iArn interface{}) {
	if iArn == nil {
		return
	}
	sArn := toString(iArn)
	if sArn == "" {
		return
	}
	if !strings.Contains(sArn, "arn:") {
		// TODO: remove printing
		fmt.Fprintf(os.Stderr, "resource %+v tried adding relationsip that was not an ARN: %s - %s\n", r.Identifier, kind.String(), sArn)
		return
	}
	parsedArn := arn.Parse(sArn)
	r.addRelation(r.Account, parsedArn.Region, kind, parsedArn.ResourceId, parsedArn.ResourceVersion)
}

func (r *Resource) AddRelation(kind ResourceType, iId, iVersion interface{}) {
	r.addRelation(r.Account, r.Region, kind, iId, iVersion)
}

func (r *Resource) AddCrossRelation(account string, iRegion interface{}, kind ResourceType, iId, iVersion interface{}) {
	region := toString(iRegion)
	if region == "" {
		region = r.Region
	}
	r.addRelation(account, region, kind, iId, iVersion)
}

func (r *Resource) addRelation(account string, region string, kind ResourceType, iId, iVersion interface{}) {
	id := toString(iId)
	version := toString(iVersion)

	if id == "" {
		return // no op if no id to relate to
	}

	if strings.Contains(id, "arn:") {
		// TODO: remove printing
		fmt.Fprintf(os.Stderr, "new relation with %s has arn: %s - %s\n", r.Type.String(), kind.String(), id)
	}
	if strings.HasPrefix(kind.String(), "iam/") ||
		strings.HasPrefix(kind.String(), "route53/") ||
		strings.HasPrefix(kind.String(), "waf/") {
		region = "aws-global"
	}
	r.Relations = append(r.Relations, Identifier{
		Account: account,
		Region:  region,
		Id:      id,
		Version: version,
		Type:    kind,
	})
}

func toString(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case *string:
		if v == nil {
			return ""
		}
		return *v
	default: // handles nil
		return ""
	}
}

func (r *Resource) JSON() (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	return string(b), err
}

func (r *Resource) YAML() (string, error) {
	b, err := yaml.Marshal(r)
	return string(b), err
}

func NewGroup() *Group {
	return &Group{
		Resources: make(map[Identifier]Resource),
	}
}

func (g *Group) Merge(group *Group) {
	if group == nil {
		return
	}
	g.Lock()
	defer g.Unlock()
	for _, res := range group.Resources {
		g.AddResource(res)
	}
}

func (g *Group) AddResource(resource Resource) {
	g.Resources[resource.Identifier] = resource
}

func (g *Group) JSON() (string, error) {
	res := make([]Resource, 0)
	for _, v := range g.Resources {
		res = append(res, v)
	}
	sort.Slice(res, func(i, j int) bool {
		a, b := res[i], res[j]

		// Sort resources in JSON by Account, Type, Region, Id, then Version to allow for consisting diff-ing

		if a.Account != b.Account {
			return a.Account < b.Account
		} else if a.Type != b.Type {
			return a.Type.String() < b.Type.String()
		} else if a.Region != b.Region {
			return a.Region < b.Region
		} else if a.Id != b.Id {
			return a.Id < b.Id
		} else {
			return a.Version < b.Version
		}
	})
	b, err := json.MarshalIndent(res, "", "  ")
	return string(b), err
}
