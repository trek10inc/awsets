package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AWSEc2InternetGateway struct {
}

func init() {
	i := AWSEc2InternetGateway{}
	listers = append(listers, i)
}

func (l AWSEc2InternetGateway) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2InternetGateway}
}

func (l AWSEc2InternetGateway) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeInternetGatewaysRequest(&ec2.DescribeInternetGatewaysInput{
		MaxResults: aws.Int64(1000),
	})

	rg := resource.NewGroup()
	paginator := ec2.NewDescribeInternetGatewaysPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.InternetGateways {
			r := resource.New(ctx, resource.Ec2InternetGateway, v.InternetGatewayId, v.InternetGatewayId, v)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
