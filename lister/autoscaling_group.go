package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
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

func (l AWSAutoscalingGroup) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := autoscaling.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeAutoScalingGroups(cfg.Context, &autoscaling.DescribeAutoScalingGroupsInput{
			MaxRecords: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.AutoScalingGroups {
			groupArn := arn.ParseP(v.AutoScalingGroupARN)
			r := resource.New(cfg, resource.AutoscalingGroup, groupArn.ResourceId, v.AutoScalingGroupName, v)

			for _, i := range v.Instances {
				r.AddRelation(resource.Ec2Instance, i.InstanceId, "")
			}

			hooks, err := svc.DescribeLifecycleHooks(cfg.Context, &autoscaling.DescribeLifecycleHooksInput{
				AutoScalingGroupName: v.AutoScalingGroupName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe lifecycle hooks for group %s: %w", *v.AutoScalingGroupName, err)
			}
			for _, hook := range hooks.LifecycleHooks {
				hookR := resource.New(cfg, resource.AutoscalingLifecycleHook, hook.LifecycleHookName, hook.LifecycleHookName, hook)
				hookR.AddRelation(resource.AutoscalingGroup, v.AutoScalingGroupName, "")
				rg.AddResource(hookR)
			}
			rg.AddResource(r)

			err = Paginator(func(nt2 *string) (*string, error) {
				scheduledActionsRes, err := svc.DescribeScheduledActions(cfg.Context, &autoscaling.DescribeScheduledActionsInput{
					AutoScalingGroupName: v.AutoScalingGroupName,
					MaxRecords:           aws.Int32(100),
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get scheduled actions for group %s: %w", *v.AutoScalingGroupName, err)
				}

				for _, action := range scheduledActionsRes.ScheduledUpdateGroupActions {
					actionR := resource.New(cfg, resource.AutoscalingScheduledAction, action.ScheduledActionName, action.ScheduledActionName, action)
					actionR.AddRelation(resource.AutoscalingGroup, v.AutoScalingGroupName, "")
					rg.AddResource(actionR)
				}
				return scheduledActionsRes.NextToken, nil
			})
			if err != nil {
				return nil, err
			}
		}

		return res.NextToken, nil
	})
	return rg, err
}
