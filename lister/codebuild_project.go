package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
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

func (l AWSCodebuildProject) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := codebuild.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListProjects(cfg.Context, &codebuild.ListProjectsInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list codebuild projects: %w", err)
		}
		if len(res.Projects) == 0 {
			return nil, nil
		}
		projects, err := svc.BatchGetProjects(cfg.Context, &codebuild.BatchGetProjectsInput{
			Names: res.Projects,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to batch get projects: %w", err)
		}
		for _, p := range projects.Projects {
			projectArn := arn.ParseP(p.Arn)
			r := resource.New(cfg, resource.CodeBuildProject, projectArn.ResourceId, p.Name, p)
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
		return res.NextToken, nil
	})
	return rg, err
}
