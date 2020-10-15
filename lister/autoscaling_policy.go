package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
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

func (l AWSAutoscalingPolicies) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := autoscaling.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribePolicies(cfg.Context, &autoscaling.DescribePoliciesInput{
			MaxRecords: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.ScalingPolicies {
			policyArn := arn.ParseP(v.PolicyARN)
			r := resource.New(cfg, resource.AutoscalingPolicy, policyArn.ResourceId, v.PolicyName, v)
			r.AddRelation(resource.AutoscalingGroup, v.AutoScalingGroupName, "")
			//TODO relation to autoscaling alarms?
			rg.AddResource(r)
		}

		return res.NextToken, nil
	})

	return rg, err
}
