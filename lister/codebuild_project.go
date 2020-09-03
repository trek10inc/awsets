package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/resource"
)

type AWSCodebuildProject struct {
}

func init() {
	i := AWSCodebuildProject{}
	listers = append(listers, i)
}

func (l AWSCodebuildProject) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CodeBuildProject}
}

func (l AWSCodebuildProject) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := codebuild.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		projectNames, err := svc.ListProjectsRequest(&codebuild.ListProjectsInput{
			NextToken: nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list codebuild projects: %w", err)
		}
		if len(projectNames.Projects) == 0 {
			break
		}
		projects, err := svc.BatchGetProjectsRequest(&codebuild.BatchGetProjectsInput{
			Names: projectNames.Projects,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to batch get projects: %w", err)
		}
		for _, p := range projects.Projects {
			projectArn := arn.ParseP(p.Arn)
			r := resource.New(ctx, resource.CodeBuildProject, projectArn.ResourceId, p.Name, p)
			if p.VpcConfig != nil {
				r.AddRelation(resource.Ec2Vpc, p.VpcConfig.VpcId, "")
				for _, sn := range p.VpcConfig.Subnets {
					r.AddRelation(resource.Ec2Subnet, sn, "")
				}
				for _, sg := range p.VpcConfig.SecurityGroupIds {
					r.AddRelation(resource.Ec2SecurityGroup, sg, "")
				}
			}
			rg.AddResource(r)
		}

		if projectNames.NextToken == nil {
			break
		}
		nextToken = projectNames.NextToken
	}
	return rg, nil
}
