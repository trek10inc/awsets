package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSDocDBSubnetGroup struct {
}

func init() {
	i := AWSDocDBSubnetGroup{}
	listers = append(listers, i)
}

func (l AWSDocDBSubnetGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DocDBSubnetGroup}
}

func (l AWSDocDBSubnetGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := docdb.New(ctx.AWSCfg)

	req := svc.DescribeDBSubnetGroupsRequest(&docdb.DescribeDBSubnetGroupsInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := docdb.NewDescribeDBSubnetGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, sg := range page.DBSubnetGroups {
			r := resource.New(ctx, resource.DocDBSubnetGroup, sg.DBSubnetGroupName, sg.DBSubnetGroupName, sg)
			r.AddRelation(resource.Ec2Vpc, sg.VpcId, "")
			for _, sn := range sg.Subnets {
				r.AddRelation(resource.Ec2Subnet, sn.SubnetIdentifier, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
