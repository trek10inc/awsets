package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2RouteTable struct {
}

func init() {
	i := AWSEc2RouteTable{}
	listers = append(listers, i)
}

func (l AWSEc2RouteTable) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2RouteTable}
}

func (l AWSEc2RouteTable) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeRouteTables(ctx.Context, &ec2.DescribeRouteTablesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.RouteTables {
			r := resource.New(ctx, resource.Ec2RouteTable, v.RouteTableId, v.RouteTableId, v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			for _, a := range v.Associations {
				r.AddRelation(resource.Ec2Subnet, a.SubnetId, "")
			}
			for _, route := range v.Routes {
				r.AddRelation(resource.Ec2Instance, route.InstanceId, "")
				r.AddRelation(resource.Ec2NatGateway, route.NatGatewayId, "")
				r.AddRelation(resource.Ec2VpcPeering, route.VpcPeeringConnectionId, "")
				r.AddRelation(resource.Ec2TransitGateway, route.TransitGatewayId, "")
				r.AddRelation(resource.Ec2NetworkInterface, route.NetworkInterfaceId, "")
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
