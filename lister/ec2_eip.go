package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2Eip struct {
}

func init() {
	i := AWSEc2Eip{}
	listers = append(listers, i)
}

func (l AWSEc2Eip) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2Eip}
}

func (l AWSEc2Eip) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	res, err := svc.DescribeAddresses(ctx.Context, &ec2.DescribeAddressesInput{})
	if err != nil {
		return rg, err
	}

	for _, v := range res.Addresses {
		r := resource.New(ctx, resource.Ec2Eip, v.PublicIp, v.PublicIp, v)
		r.AddRelation(resource.Ec2Instance, v.InstanceId, "")
		rg.AddResource(r)
	}
	return rg, nil
}
