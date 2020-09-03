package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
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
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeRouteTablesRequest(&ec2.DescribeRouteTablesInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := ec2.NewDescribeRouteTablesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.RouteTables {
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
	}
	err := paginator.Err()
	return rg, err
}
