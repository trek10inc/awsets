package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloud9"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSCloud9Environment struct {
}

func init() {
	i := AWSCloud9Environment{}
	listers = append(listers, i)
}

func (l AWSCloud9Environment) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Cloud9Environment}
}

func (l AWSCloud9Environment) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloud9.New(ctx.AWSCfg)

	req := svc.ListEnvironmentsRequest(&cloud9.ListEnvironmentsInput{
		MaxResults: aws.Int64(25),
	})

	rg := resource.NewGroup()
	paginator := cloud9.NewListEnvironmentsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		if len(page.EnvironmentIds) == 0 {
			continue
		}
		environments, err := svc.DescribeEnvironmentsRequest(&cloud9.DescribeEnvironmentsInput{
			EnvironmentIds: page.EnvironmentIds,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to describe environments: %w", err)
		}
		for _, v := range environments.Environments {
			r := resource.New(ctx, resource.Cloud9Environment, v.Name, v.Name, v)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
