package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AWSEc2Vpc struct {
}

func init() {
	i := AWSEc2Vpc{}
	listers = append(listers, i)
}

func (l AWSEc2Vpc) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2Vpc}
}

func (l AWSEc2Vpc) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeVpcsRequest(&ec2.DescribeVpcsInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := ec2.NewDescribeVpcsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Vpcs {
			r := resource.New(ctx, resource.Ec2Vpc, v.VpcId, v.VpcId, v)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
