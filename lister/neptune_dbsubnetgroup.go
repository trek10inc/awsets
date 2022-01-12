package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSNeptuneDbSubnetGroup struct {
}

func init() {
	i := AWSNeptuneDbSubnetGroup{}
	listers = append(listers, i)
}

func (l AWSNeptuneDbSubnetGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.NeptuneDbSubnetGroup}
}

func (l AWSNeptuneDbSubnetGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := neptune.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	paginator := neptune.NewDescribeDBSubnetGroupsPaginator(svc, &neptune.DescribeDBSubnetGroupsInput{
		MaxRecords: aws.Int32(100),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx.Context)
		if err != nil {
			return nil, err
		}
		for _, v := range page.DBSubnetGroups {
			subnetArn := arn.ParseP(v.DBSubnetGroupArn)
			if subnetArn.Service != "neptune" {
				continue
			}
			r := resource.New(ctx, resource.NeptuneDbSubnetGroup, subnetArn.ResourceId, "", v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			for _, subnet := range v.Subnets {
				r.AddRelation(resource.Ec2Subnet, subnet.SubnetIdentifier, "")
			}
			rg.AddResource(r)
		}
	}
	return rg, nil
}
