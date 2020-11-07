package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2VpnConnection struct {
}

func init() {
	i := AWSEc2VpnConnection{}
	listers = append(listers, i)
}

func (l AWSEc2VpnConnection) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.Ec2VpnConnection,
	}
}

func (l AWSEc2VpnConnection) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ec2.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	res, err := svc.DescribeVpnConnections(cfg.Context, &ec2.DescribeVpnConnectionsInput{
		//VpnConnectionIds: nil,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get vpn connections: %w, err")
	}
	for _, v := range res.VpnConnections {
		r := resource.New(cfg, resource.Ec2VpnConnection, v.VpnConnectionId, v.VpnConnectionId, v)
		r.AddRelation(resource.Ec2CustomerGateway, v.CustomerGatewayId, "")
		r.AddRelation(resource.Ec2TransitGateway, v.TransitGatewayId, "")
		r.AddRelation(resource.Ec2VpnGateway, v.VpnGatewayId, "")
		rg.AddResource(r)
	}

	return rg, err
}
