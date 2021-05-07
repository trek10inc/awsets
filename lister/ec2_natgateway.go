package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := ec2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeNatGateways(ctx.Context, &ec2.DescribeNatGatewaysInput{
			MaxResults: aws.Int32(1000),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.NatGateways {
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
		return res.NextToken, nil
	})
	return rg, err
}
