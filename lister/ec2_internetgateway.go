package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := ec2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {

		res, err := svc.DescribeInternetGateways(ctx.Context, &ec2.DescribeInternetGatewaysInput{
			MaxResults: aws.Int32(1000),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.InternetGateways {
			r := resource.New(ctx, resource.Ec2InternetGateway, v.InternetGatewayId, v.InternetGatewayId, v)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
