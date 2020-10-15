package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSWafRegionalRule struct {
}

func init() {
	i := AWSWafRegionalRule{}
	listers = append(listers, i)
}

func (l AWSWafRegionalRule) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafRegionalRule}
}

func (l AWSWafRegionalRule) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListRules(cfg.Context, &wafregional.ListRulesInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list webacls: %w", err)
		}
		for _, ruleId := range res.Rules {
			rule, err := svc.GetRule(cfg.Context, &wafregional.GetRuleInput{RuleId: ruleId.RuleId})
			if err != nil {
				return nil, fmt.Errorf("failed to get rule %s: %w", *ruleId.RuleId, err)
			}
			if rule.Rule == nil {
				continue
			}
			r := resource.New(cfg, resource.WafRegionalRule, rule.Rule.RuleId, rule.Rule.Name, rule.Rule)
			rg.AddResource(r)
		}
		return res.NextMarker, nil
	})
	return rg, err
}
