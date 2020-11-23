package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSIoTTopicRule struct {
}

func init() {
	i := AWSIoTTopicRule{}
	listers = append(listers, i)
}

func (l AWSIoTTopicRule) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.IotTopicRule}
}

func (l AWSIoTTopicRule) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := iot.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListTopicRules(ctx.Context, &iot.ListTopicRulesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list iot topic rules: %w", err)
		}
		for _, rule := range res.Rules {
			ruleArn := arn.ParseP(rule.RuleArn)
			r := resource.New(ctx, resource.IotTopicRule, ruleArn.ResourceId, rule.RuleName, rule)

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
