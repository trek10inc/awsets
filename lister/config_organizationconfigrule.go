package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/trek10inc/awsets/arn"
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

	svc := configservice.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeOrganizationConfigRules(ctx.Context, &configservice.DescribeOrganizationConfigRulesInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list config organization config rules: %w", err)
		}
		for _, v := range res.OrganizationConfigRules {
			ruleArn := arn.ParseP(v.OrganizationConfigRuleArn)
			r := resource.New(ctx, resource.ConfigOrganizationConfigRule, ruleArn.ResourceId, v.OrganizationConfigRuleName, v)
			if v.OrganizationCustomRuleMetadata != nil {
				r.AddARNRelation(resource.LambdaFunction, v.OrganizationCustomRuleMetadata.LambdaFunctionArn)
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
