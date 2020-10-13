package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/redshift"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSRedshiftParameterGroup struct {
}

func init() {
	i := AWSRedshiftParameterGroup{}
	listers = append(listers, i)
}

func (l AWSRedshiftParameterGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.RedshiftParameterGroup}
}

func (l AWSRedshiftParameterGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := redshift.NewFromConfig(ctx.AWSCfg)

	res, err := svc.DescribeClusterParameterGroups(ctx.Context, &redshift.DescribeClusterParameterGroupsInput{
		MaxRecords: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := redshift.NewDescribeClusterParameterGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, pg := range page.ParameterGroups {
			r := resource.New(ctx, resource.RedshiftParameterGroup, pg.ParameterGroupName, pg.ParameterGroupName, pg)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
