package lister

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
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

func (l AWSEc2TransitGateway) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeTransitGatewaysRequest(&ec2.DescribeTransitGatewaysInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := ec2.NewDescribeTransitGatewaysPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.TransitGateways {
			r := resource.New(ctx, resource.Ec2TransitGateway, v.TransitGatewayId, v.TransitGatewayId, v)
			// TODO lots of additional info to query here
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "InvalidAction" &&
				strings.Contains(aerr.Message(), "is not valid for this web service") {
				// If transit gateways are not supported in a region, returns invalid action
				err = nil
			}
		}
	}
	return rg, err
}
