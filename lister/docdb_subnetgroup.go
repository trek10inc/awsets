package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	"github.com/trek10inc/awsets/arn"
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
	svc := docdb.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	paginator := docdb.NewDescribeDBSubnetGroupsPaginator(svc, &docdb.DescribeDBSubnetGroupsInput{
		MaxRecords: aws.Int32(100),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx.Context)
		if err != nil {
			return nil, err
		}
		for _, v := range page.DBSubnetGroups {
			subnetArn := arn.ParseP(v.DBSubnetGroupArn)
			if subnetArn.Service != "docdb" {
				continue
			}
			r := resource.New(ctx, resource.DocDBSubnetGroup, v.DBSubnetGroupName, v.DBSubnetGroupName, v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			for _, sn := range v.Subnets {
				r.AddRelation(resource.Ec2Subnet, sn.SubnetIdentifier, "")
			}
			rg.AddResource(r)
		}
	}
	return rg, nil
}
