package lister

import (
	"fmt"
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	"github.com/trek10inc/awsets/resource"
)

var listWafRulesOnce sync.Once

type AWSWafRule struct {
}

func init() {
	i := AWSWafRule{}
	listers = append(listers, i)
}

func (l AWSWafRule) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafRule}
}

func (l AWSWafRule) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := waf.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	var outerErr error

	listWafRulesOnce.Do(func() {
		var nextMarker *string
		for {
			res, err := svc.ListRules(ctx.Context, &waf.ListRulesInput{
				Limit:      aws.Int32(100),
				NextMarker: nextMarker,
			})
			if err != nil {
				outerErr = fmt.Errorf("failed to list webacls: %w", err)
				return
			}
			for _, ruleId := range res.Rules {
				rule, err := svc.GetRule(ctx.Context, &waf.GetRuleInput{RuleId: ruleId.RuleId})
				if err != nil {
					outerErr = fmt.Errorf("failed to get rule %s: %w", aws.StringValue(ruleId.RuleId), err)
					return
				}
				if rule.Rule == nil {
					continue
				}
				r := resource.NewGlobal(ctx, resource.WafRule, rule.Rule.RuleId, rule.Rule.Name, rule.Rule)
				rg.AddResource(r)
			}
			if res.NextMarker == nil {
				break
			}
			nextMarker = res.NextMarker
		}
	})

	return rg, outerErr
}
