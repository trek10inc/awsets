package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSRdsDbSubnetGroup struct {
}

func init() {
	i := AWSRdsDbSubnetGroup{}
	listers = append(listers, i)
}

func (l AWSRdsDbSubnetGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.RdsDbSubnetGroup}
}

func (l AWSRdsDbSubnetGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := rds.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBSubnetGroups(ctx.Context, &rds.DescribeDBSubnetGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, subnetGroup := range res.DBSubnetGroups {
			subnetArn := arn.ParseP(subnetGroup.DBSubnetGroupArn)
			r := resource.New(ctx, resource.RdsDbSubnetGroup, subnetArn.ResourceId, "", subnetGroup)
			r.AddRelation(resource.Ec2Vpc, subnetGroup.VpcId, "")
			for _, subnet := range subnetGroup.Subnets {
				r.AddRelation(resource.Ec2Subnet, subnet.SubnetIdentifier, "")
			}
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
