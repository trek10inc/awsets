package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2DHCPOption struct {
}

func init() {
	i := AWSEc2DHCPOption{}
	listers = append(listers, i)
}

func (l AWSEc2DHCPOption) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.Ec2DHCPOption,
	}
}

func (l AWSEc2DHCPOption) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDhcpOptions(ctx.Context, &ec2.DescribeDhcpOptionsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.DhcpOptions {
			r := resource.New(ctx, resource.Ec2DHCPOption, v.DhcpOptionsId, v.DhcpOptionsId, v)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
