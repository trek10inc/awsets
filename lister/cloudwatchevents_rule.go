package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchevents"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSCloudwatchEventsRule struct {
}

func init() {
	i := AWSCloudwatchEventsRule{}
	listers = append(listers, i)
}

func (l AWSCloudwatchEventsRule) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.EventsRule}
}

func (l AWSCloudwatchEventsRule) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := cloudwatchevents.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListRules(ctx.Context, &cloudwatchevents.ListRulesInput{
			Limit:     aws.Int32(100),
			NextToken: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list cloudwatch event rules: %w", err)
		}
		for _, rule := range res.Rules {
			r := resource.New(ctx, resource.EventsRule, rule.Name, rule.Name, rule)
			r.AddRelation(resource.EventsBus, rule.EventBusName, "")
			if rule.RoleArn != nil {
				roleArn := arn.ParseP(rule.RoleArn)
				r.AddRelation(resource.IamRole, roleArn.ResourceId, roleArn.ResourceVersion)
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
