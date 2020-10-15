package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

var listWafv2WebACLOnce sync.Once

type AWSWafv2WebACL struct {
}

func init() {
	i := AWSWafv2WebACL{}
	listers = append(listers, i)
}

func (l AWSWafv2WebACL) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Wafv2WebACL}
}

func (l AWSWafv2WebACL) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	rg := resource.NewGroup()

	rg, err := wafv2WebACLQuery(cfg, types.ScopeRegional)
	if err != nil {
		return nil, fmt.Errorf("failed to list web ACLs: %w", err)
	}

	// Do global
	var outerErr error
	listWafv2WebACLOnce.Do(func() {
		ctxUsEast := cfg.Copy("us-east-1")
		rgNew, err := wafv2WebACLQuery(ctxUsEast, types.ScopeCloudfront)
		if err != nil {
			outerErr = fmt.Errorf("failed to list global web ACLs: %w", err)
		}
		rg.Merge(rgNew)
	})

	return rg, outerErr
}

func wafv2WebACLQuery(cfg option.AWSetsConfig, scope types.Scope) (*resource.Group, error) {
	svc := wafv2.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListWebACLs(cfg.Context, &wafv2.ListWebACLsInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
			Scope:      scope,
		})
		if err != nil {
			return nil, err
		}
		for _, waId := range res.WebACLs {
			wa, err := svc.GetWebACL(cfg.Context, &wafv2.GetWebACLInput{
				Id:    waId.Id,
				Name:  waId.Name,
				Scope: scope,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get web ACL %s for scope %v: %w", *waId.Id, scope, err)
			}
			if v := wa.WebACL; v != nil {
				var r resource.Resource
				if scope == types.ScopeCloudfront {
					r = resource.NewGlobal(cfg, resource.Wafv2WebACL, v.Id, v.Name, v)
				} else {
					r = resource.New(cfg, resource.Wafv2WebACL, v.Id, v.Name, v)
				}
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
