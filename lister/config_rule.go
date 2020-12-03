package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/trek10inc/awsets/arn"
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
	return []resource.ResourceType{
		resource.ConfigRule,
		resource.ConfigRemediationConfiguration,
	}
}

func (l AWSConfigRule) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := configservice.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeConfigRules(ctx.Context, &configservice.DescribeConfigRulesInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list config rules: %w", err)
		}

		ruleNames := make([]string, 0)
		for _, rule := range res.ConfigRules {
			r := resource.New(ctx, resource.ConfigRule, rule.ConfigRuleId, rule.ConfigRuleName, rule)
			rg.AddResource(r)
			ruleNames = append(ruleNames, *rule.ConfigRuleName)
		}

		// Remediation Configs for the rules
		remediationConfigs, err := svc.DescribeRemediationConfigurations(ctx.Context, &configservice.DescribeRemediationConfigurationsInput{
			ConfigRuleNames: ruleNames,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list remediation configurations: %w", err)
		}
		for _, v := range remediationConfigs.RemediationConfigurations {
			configArn := arn.ParseP(v.Arn)
			r := resource.New(ctx, resource.ConfigRemediationConfiguration, configArn.ResourceId, configArn.ResourceId, v)
			r.AddRelation(resource.ConfigRule, v.ConfigRuleName, "")
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
