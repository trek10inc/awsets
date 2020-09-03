package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSAutoscalingLaunchConfiguration struct {
}

func init() {
	i := AWSAutoscalingLaunchConfiguration{}
	listers = append(listers, i)
}

func (l AWSAutoscalingLaunchConfiguration) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.AutoscalingLaunchConfig}
}

func (l AWSAutoscalingLaunchConfiguration) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := autoscaling.New(ctx.AWSCfg)
	req := svc.DescribeLaunchConfigurationsRequest(&autoscaling.DescribeLaunchConfigurationsInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := autoscaling.NewDescribeLaunchConfigurationsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.LaunchConfigurations {

			configArn := arn.ParseP(v.LaunchConfigurationARN)
			r := resource.New(ctx, resource.AutoscalingLaunchConfig, configArn.ResourceId, v.LaunchConfigurationName, v)

			for _, sg := range v.SecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, sg, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
