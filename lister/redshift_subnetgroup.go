package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/redshift"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSRedshiftSubnetGroup struct {
}

func init() {
	i := AWSRedshiftSubnetGroup{}
	listers = append(listers, i)
}

func (l AWSRedshiftSubnetGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.RedshiftSubnetGroup}
}

func (l AWSRedshiftSubnetGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := redshift.NewFromConfig(ctx.AWSCfg)

	res, err := svc.DescribeClusterSubnetGroups(ctx.Context, &redshift.DescribeClusterSubnetGroupsInput{
		MaxRecords: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := redshift.NewDescribeClusterSubnetGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, sg := range page.ClusterSubnetGroups {
			r := resource.New(ctx, resource.RedshiftSubnetGroup, sg.ClusterSubnetGroupName, sg.ClusterSubnetGroupName, sg)
			r.AddRelation(resource.Ec2Vpc, sg.VpcId, "")
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
