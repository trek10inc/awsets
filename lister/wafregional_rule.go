package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
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

func (l AWSWafRegionalRule) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := wafregional.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextMarker *string
	for {
		res, err := svc.ListRulesRequest(&wafregional.ListRulesInput{
			Limit:      aws.Int64(100),
			NextMarker: nextMarker,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list webacls: %w", err)
		}
		for _, ruleId := range res.Rules {
			rule, err := svc.GetRuleRequest(&wafregional.GetRuleInput{RuleId: ruleId.RuleId}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get rule %s: %w", aws.StringValue(ruleId.RuleId), err)
			}
			if rule.Rule == nil {
				continue
			}
			r := resource.New(ctx, resource.WafRegionalRule, rule.Rule.RuleId, rule.Rule.Name, rule.Rule)
			rg.AddResource(r)
		}
		if res.NextMarker == nil {
			break
		}
		nextMarker = res.NextMarker
	}

	return rg, nil
}
