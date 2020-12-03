package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSWafRegionalRateBasedRule struct {
}

func init() {
	i := AWSWafRegionalRateBasedRule{}
	listers = append(listers, i)
}

func (l AWSWafRegionalRateBasedRule) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafRegionalRateBasedRule}
}

func (l AWSWafRegionalRateBasedRule) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListRateBasedRules(ctx.Context, &wafregional.ListRateBasedRulesInput{
			Limit:      100,
			NextMarker: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list regional rate based rule: %w", err)
		}
		for _, id := range res.Rules {
			rule, err := svc.GetRateBasedRule(ctx.Context, &wafregional.GetRateBasedRuleInput{
				RuleId: id.RuleId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get rate based rule %s: %w", *id.RuleId, err)
			}
			if v := rule.Rule; v != nil {
				r := resource.New(ctx, resource.WafRegionalRateBasedRule, v.RuleId, v.Name, v)
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
