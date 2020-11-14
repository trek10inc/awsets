package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeClusterSubnetGroups(ctx.Context, &redshift.DescribeClusterSubnetGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, sg := range res.ClusterSubnetGroups {
			r := resource.New(ctx, resource.RedshiftSubnetGroup, sg.ClusterSubnetGroupName, sg.ClusterSubnetGroupName, sg)
			r.AddRelation(resource.Ec2Vpc, sg.VpcId, "")
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
