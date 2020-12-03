package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/waf"
	"github.com/trek10inc/awsets/context"
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
		outerErr = Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListRules(ctx.Context, &waf.ListRulesInput{
				Limit:      100,
				NextMarker: nt,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list rules: %w", err)
			}
			for _, ruleId := range res.Rules {
				rule, err := svc.GetRule(ctx.Context, &waf.GetRuleInput{RuleId: ruleId.RuleId})
				if err != nil {
					return nil, fmt.Errorf("failed to get rule %s: %w", *ruleId.RuleId, err)
				}
				if rule.Rule == nil {
					continue
				}
				r := resource.NewGlobal(ctx, resource.WafRule, rule.Rule.RuleId, rule.Rule.Name, rule.Rule)
				rg.AddResource(r)
			}
			return res.NextMarker, nil
		})
	})

	return rg, outerErr
}
