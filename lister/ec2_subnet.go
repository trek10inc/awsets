package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AWSEc2Subnet struct {
}

func init() {
	i := AWSEc2Subnet{}
	listers = append(listers, i)
}

func (l AWSEc2Subnet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2Subnet}
}

func (l AWSEc2Subnet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeSubnetsRequest(&ec2.DescribeSubnetsInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := ec2.NewDescribeSubnetsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Subnets {
			r := resource.New(ctx, resource.Ec2Subnet, v.SubnetId, v.SubnetArn, v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
