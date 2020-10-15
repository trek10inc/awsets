package lister

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2TransitGateway struct {
}

func init() {
	i := AWSEc2TransitGateway{}
	listers = append(listers, i)
}

func (l AWSEc2TransitGateway) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2TransitGateway}
}

func (l AWSEc2TransitGateway) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ec2.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeTransitGateways(cfg.Context, &ec2.DescribeTransitGatewaysInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "is not valid for this web service") {
				// If transit gateways are not supported in a region, returns invalid action
				return nil, nil
			}
			return nil, err
		}
		for _, v := range res.TransitGateways {
			r := resource.New(cfg, resource.Ec2TransitGateway, v.TransitGatewayId, v.TransitGatewayId, v)
			// TODO lots of additional info to query here
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
