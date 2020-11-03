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

var listWafv2RuleGroupOnce sync.Once

type AWSWafv2RuleGroup struct {
}

func init() {
	i := AWSWafv2RuleGroup{}
	listers = append(listers, i)
}

func (l AWSWafv2RuleGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Wafv2RuleGroup}
}

func (l AWSWafv2RuleGroup) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	rg := resource.NewGroup()

	rg, err := wafv2RuleGroupQuery(cfg, types.ScopeRegional)
	if err != nil {
		return nil, fmt.Errorf("failed to list rule groups: %w", err)
	}

	// Do global
	var outerErr error
	listWafv2RuleGroupOnce.Do(func() {
		ctxUsEast := cfg.CopyWithRegion("us-east-1")
		rgNew, err := wafv2RuleGroupQuery(*ctxUsEast, types.ScopeCloudfront)
		if err != nil {
			outerErr = fmt.Errorf("failed to list global rule groups: %w", err)
		}
		rg.Merge(rgNew)
	})

	return rg, outerErr
}

func wafv2RuleGroupQuery(cfg option.AWSetsConfig, scope types.Scope) (*resource.Group, error) {
	svc := wafv2.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListRuleGroups(cfg.Context, &wafv2.ListRuleGroupsInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
			Scope:      scope,
		})
		if err != nil {
			return nil, err
		}
		for _, rgId := range res.RuleGroups {
			rulegroup, err := svc.GetRuleGroup(cfg.Context, &wafv2.GetRuleGroupInput{
				Id:    rgId.Id,
				Name:  rgId.Name,
				Scope: scope,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get rule group %s for scope %v: %w", *rgId.Id, scope, err)
			}
			if v := rulegroup.RuleGroup; v != nil {
				var r resource.Resource
				if scope == types.ScopeCloudfront {
					r = resource.NewGlobal(cfg, resource.Wafv2RuleGroup, v.Id, v.Name, v)
				} else {
					r = resource.New(cfg, resource.Wafv2RuleGroup, v.Id, v.Name, v)
				}
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
