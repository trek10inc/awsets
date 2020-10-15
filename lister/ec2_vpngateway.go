package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2VpnGateway struct {
}

func init() {
	i := AWSEc2VpnGateway{}
	listers = append(listers, i)
}

func (l AWSEc2VpnGateway) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2VpnGateway}
}

func (l AWSEc2VpnGateway) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ec2.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()

	gateways, err := svc.DescribeVpnGateways(cfg.Context, &ec2.DescribeVpnGatewaysInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list vpn gateways: %w", err)
	}

	for _, v := range gateways.VpnGateways {
		r := resource.New(cfg, resource.Ec2VpnGateway, v.VpnGatewayId, v.VpnGatewayId, v)
		for _, a := range v.VpcAttachments {
			r.AddRelation(resource.Ec2Vpc, a.VpcId, "")
		}
		rg.AddResource(r)
	}

	return rg, nil
}
