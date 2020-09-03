package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSAutoscalingPolicies struct {
}

func init() {
	i := AWSAutoscalingPolicies{}
	listers = append(listers, i)
}

func (l AWSAutoscalingPolicies) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.AutoscalingPolicy}
}

func (l AWSAutoscalingPolicies) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := autoscaling.New(ctx.AWSCfg)

	req := svc.DescribePoliciesRequest(&autoscaling.DescribePoliciesInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := autoscaling.NewDescribePoliciesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.ScalingPolicies {
			policyArn := arn.ParseP(v.PolicyARN)
			r := resource.New(ctx, resource.AutoscalingPolicy, policyArn.ResourceId, v.PolicyName, v)
			r.AddRelation(resource.AutoscalingGroup, v.AutoScalingGroupName, "")
			//TODO relation to autoscaling alarms?
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
