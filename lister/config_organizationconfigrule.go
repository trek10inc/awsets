package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/service/configservice"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"
)

type AWSConfigOrganizationConfigRule struct {
}

func init() {
	i := AWSConfigOrganizationConfigRule{}
	listers = append(listers, i)
}

func (l AWSConfigOrganizationConfigRule) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ConfigOrganizationConfigRule}
}

func (l AWSConfigOrganizationConfigRule) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := configservice.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextToken *string
	for {
		rules, err := svc.DescribeOrganizationConfigRulesRequest(&configservice.DescribeOrganizationConfigRulesInput{
			NextToken: nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list config organization config rules: %w", err)
		}
		for _, v := range rules.OrganizationConfigRules {
			ruleArn := arn.ParseP(v.OrganizationConfigRuleArn)
			r := resource.New(ctx, resource.ConfigOrganizationConfigRule, ruleArn.ResourceId, v.OrganizationConfigRuleName, v)
			if v.OrganizationCustomRuleMetadata != nil {
				r.AddARNRelation(resource.LambdaFunction, v.OrganizationCustomRuleMetadata.LambdaFunctionArn)
			}
			rg.AddResource(r)
		}
		if rules.NextToken == nil {
			break
		}
		nextToken = rules.NextToken
	}
	return rg, nil
}
