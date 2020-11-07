package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2CustomerGateway struct {
}

func init() {
	i := AWSEc2CustomerGateway{}
	listers = append(listers, i)
}

func (l AWSEc2CustomerGateway) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2CustomerGateway}
}

func (l AWSEc2CustomerGateway) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ec2.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	res, err := svc.DescribeCustomerGateways(cfg.Context, &ec2.DescribeCustomerGatewaysInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get customer gateways: %w", err)
	}
	for _, v := range res.CustomerGateways {
		r := resource.New(cfg, resource.Ec2CustomerGateway, v.CustomerGatewayId, v.CustomerGatewayId, v)
		rg.AddResource(r)
	}
	return rg, err
}
