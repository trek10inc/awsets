package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/service/iot"
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

	svc := iot.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		topicRules, err := svc.ListTopicRulesRequest(&iot.ListTopicRulesInput{
			MaxResults: aws.Int64(100),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list iot topic rules: %w", err)
		}
		for _, rule := range topicRules.Rules {
			ruleArn := arn.ParseP(rule.RuleArn)
			r := resource.New(ctx, resource.IotTopicRule, ruleArn.ResourceId, rule.RuleName, rule)

			rg.AddResource(r)
		}
		if topicRules.NextToken == nil {
			break
		}
		nextToken = topicRules.NextToken
	}
	return rg, nil
}
