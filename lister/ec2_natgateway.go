package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AWSEc2NatGateway struct {
}

func init() {
	i := AWSEc2NatGateway{}
	listers = append(listers, i)
}

func (l AWSEc2NatGateway) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2NatGateway}
}

func (l AWSEc2NatGateway) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeNatGatewaysRequest(&ec2.DescribeNatGatewaysInput{
		MaxResults: aws.Int64(1000),
	})

	rg := resource.NewGroup()
	paginator := ec2.NewDescribeNatGatewaysPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.NatGateways {
			r := resource.New(ctx, resource.Ec2NatGateway, v.NatGatewayId, v.NatGatewayId, v)
			if v.SubnetId != nil {
				r.AddRelation(resource.Ec2Subnet, v.SubnetId, "")
			}
			if v.VpcId != nil {
				r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			}
			for _, gwAddress := range v.NatGatewayAddresses {
				r.AddRelation(resource.Ec2NetworkInterface, gwAddress.NetworkInterfaceId, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
