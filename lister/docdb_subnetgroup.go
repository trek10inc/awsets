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
	svc := docdb.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBSubnetGroups(ctx.Context, &docdb.DescribeDBSubnetGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, sg := range res.DBSubnetGroups {
			r := resource.New(ctx, resource.DocDBSubnetGroup, sg.DBSubnetGroupName, sg.DBSubnetGroupName, sg)
			r.AddRelation(resource.Ec2Vpc, sg.VpcId, "")
			for _, sn := range sg.Subnets {
				r.AddRelation(resource.Ec2Subnet, sn.SubnetIdentifier, "")
			}
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
