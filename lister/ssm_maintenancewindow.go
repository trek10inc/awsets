package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSSsmMaintenanceWindow struct {
}

func init() {
	i := AWSSsmMaintenanceWindow{}
	listers = append(listers, i)
}

func (l AWSSsmMaintenanceWindow) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.SsmMaintenanceWindow,
		resource.SsmMaintenanceWindowTask,
	}
}

func (l AWSSsmMaintenanceWindow) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ssm.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeMaintenanceWindows(cfg.Context, &ssm.DescribeMaintenanceWindowsInput{
			MaxResults: aws.Int32(50),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, wi := range res.WindowIdentities {
			v, err := svc.GetMaintenanceWindow(cfg.Context, &ssm.GetMaintenanceWindowInput{
				WindowId: wi.WindowId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get maintenance window %s: %w", *wi.WindowId, err)
			}

			r := resource.New(cfg, resource.SsmMaintenanceWindow, v.WindowId, v.Name, v)

			// Tasks
			err = Paginator(func(nt2 *string) (*string, error) {
				tasks, err := svc.DescribeMaintenanceWindowTasks(cfg.Context, &ssm.DescribeMaintenanceWindowTasksInput{
					WindowId:   wi.WindowId,
					MaxResults: aws.Int32(100),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get maintenance window tasks for %s: %w", *wi.WindowId, err)
				}

				for _, task := range tasks.Tasks {
					taskR := resource.New(cfg, resource.SsmMaintenanceWindowTask, task.WindowTaskId, task.Name, task)
					taskR.AddRelation(resource.SsmMaintenanceWindow, wi.WindowId, "")
					taskR.AddARNRelation(resource.IamRole, task.ServiceRoleArn)
					for _, t := range task.Targets {
						if *t.Key == "instanceids" {
							for _, val := range t.Values {
								taskR.AddRelation(resource.Ec2Instance, val, "")
							}
						}
					}
					rg.AddResource(taskR)
				}

				return tasks.NextToken, nil
			})

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
