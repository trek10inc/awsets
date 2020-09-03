package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
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

func (l AWSEc2VpnGateway) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.New(ctx.AWSCfg)

	rg := resource.NewGroup()

	gateways, err := svc.DescribeVpnGatewaysRequest(&ec2.DescribeVpnGatewaysInput{}).Send(ctx.Context)
	if err != nil {
		return rg, fmt.Errorf("failed to list vpn gateways: %w", err)
	}

	for _, v := range gateways.VpnGateways {
		r := resource.New(ctx, resource.Ec2VpnGateway, v.VpnGatewayId, v.VpnGatewayId, v)
		for _, a := range v.VpcAttachments {
			r.AddRelation(resource.Ec2Vpc, a.VpcId, "")
		}
		rg.AddResource(r)
	}

	return rg, nil
}
