package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSCodeDeployApplication struct {
}

func init() {
	i := AWSCodeDeployApplication{}
	listers = append(listers, i)
}

func (l AWSCodeDeployApplication) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.CodeDeployApplication,
		resource.CodeDeployDeploymentGroup,
	}
}

func (l AWSCodeDeployApplication) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := codedeploy.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListApplications(ctx.Context, &codedeploy.ListApplicationsInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, app := range res.Applications {
			appRes, err := svc.GetApplication(ctx.Context, &codedeploy.GetApplicationInput{
				ApplicationName: app,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get codedeploy application %s: %w", *app, err)
			}
			v := appRes.Application
			if v == nil {
				continue
			}
			r := resource.New(ctx, resource.CodeDeployApplication, v.ApplicationId, v.ApplicationName, v)

			// CodeDeploy Deployment Groups
			err = Paginator(func(nt2 *string) (*string, error) {
				depGroups, err := svc.ListDeploymentGroups(ctx.Context, &codedeploy.ListDeploymentGroupsInput{
					ApplicationName: app,
					NextToken:       nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list codedeploy deployment groups for app %s: %w", *app, err)
				}
				groups := depGroups.DeploymentGroups
				if len(groups) > 0 {
					groupsRes, err := svc.BatchGetDeploymentGroups(ctx.Context, &codedeploy.BatchGetDeploymentGroupsInput{
						ApplicationName:      app,
						DeploymentGroupNames: groups,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to get codedeploy deployment groups for app %s: %w", *app, err)
					}
					for _, group := range groupsRes.DeploymentGroupsInfo {
						groupR := resource.New(ctx, resource.CodeDeployDeploymentGroup, group.DeploymentGroupId, group.DeploymentGroupName, group)
						groupR.AddRelation(resource.CodeDeployApplication, app, "")
						groupR.AddARNRelation(resource.IamRole, group.ServiceRoleArn)
						groupR.AddRelation(resource.CodeDeployDeploymentConfig, group.DeploymentConfigName, "")
						// TODO: add relationships for ECS, ELB, etc
						rg.AddResource(groupR)
					}
				}

				return depGroups.NextToken, nil
			})
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
