package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/configservice"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"
)

type AWSConfigRule struct {
}

func init() {
	i := AWSConfigRule{}
	listers = append(listers, i)
}

func (l AWSConfigRule) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ConfigRule}
}

func (l AWSConfigRule) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := configservice.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextToken *string

	for {
		rules, err := svc.DescribeConfigRulesRequest(&configservice.DescribeConfigRulesInput{
			NextToken: nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list config rules: %w", err)
		}
		for _, rule := range rules.ConfigRules {
			r := resource.New(ctx, resource.ConfigRule, rule.ConfigRuleId, rule.ConfigRuleName, rule)
			rg.AddResource(r)
		}
		if rules.NextToken == nil {
			break
		}
		nextToken = rules.NextToken
	}

	return rg, nil
}
