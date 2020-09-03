package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSAutoscalingGroup struct {
}

func init() {
	i := AWSAutoscalingGroup{}
	listers = append(listers, i)
}

func (l AWSAutoscalingGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.AutoscalingGroup,
		resource.AutoscalingLifecycleHook,
		resource.AutoscalingScheduledAction,
	}
}

func (l AWSAutoscalingGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := autoscaling.New(ctx.AWSCfg)

	req := svc.DescribeAutoScalingGroupsRequest(&autoscaling.DescribeAutoScalingGroupsInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := autoscaling.NewDescribeAutoScalingGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.AutoScalingGroups {
			groupArn := arn.ParseP(v.AutoScalingGroupARN)
			r := resource.New(ctx, resource.AutoscalingGroup, groupArn.ResourceId, v.AutoScalingGroupName, v)

			for _, i := range v.Instances {
				r.AddRelation(resource.Ec2Instance, i.InstanceId, "")
			}

			hooks, err := svc.DescribeLifecycleHooksRequest(&autoscaling.DescribeLifecycleHooksInput{
				AutoScalingGroupName: v.AutoScalingGroupName,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to describe lifecycle hooks for group %s: %w", *v.AutoScalingGroupName, err)
			}
			for _, hook := range hooks.LifecycleHooks {
				hookR := resource.New(ctx, resource.AutoscalingLifecycleHook, hook.LifecycleHookName, hook.LifecycleHookName, hook)
				hookR.AddRelation(resource.AutoscalingGroup, v.AutoScalingGroupName, "")
				rg.AddResource(hookR)
			}
			rg.AddResource(r)

			scheduledActionsReq := svc.DescribeScheduledActionsRequest(&autoscaling.DescribeScheduledActionsInput{
				AutoScalingGroupName: v.AutoScalingGroupName,
				MaxRecords:           aws.Int64(100),
			})
			scheduledActionsPaginator := autoscaling.NewDescribeScheduledActionsPaginator(scheduledActionsReq)
			for scheduledActionsPaginator.Next(ctx.Context) {
				actionsPage := scheduledActionsPaginator.CurrentPage()
				for _, action := range actionsPage.ScheduledUpdateGroupActions {
					actionR := resource.New(ctx, resource.AutoscalingScheduledAction, action.ScheduledActionName, action.ScheduledActionName, action)
					actionR.AddRelation(resource.AutoscalingGroup, v.AutoScalingGroupName, "")
					rg.AddResource(actionR)
				}
			}
			if err := scheduledActionsPaginator.Err(); err != nil {
				return rg, fmt.Errorf("failed to get scheduled actions for group %s: %w", *v.AutoScalingGroupName, err)
			}
		}
	}
	err := paginator.Err()
	return rg, err
}
