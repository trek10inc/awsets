package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeClusterParameterGroups(ctx.Context, &redshift.DescribeClusterParameterGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, pg := range res.ParameterGroups {
			r := resource.New(ctx, resource.RedshiftParameterGroup, pg.ParameterGroupName, pg.ParameterGroupName, pg)
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
