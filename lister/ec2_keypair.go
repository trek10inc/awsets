package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2KeyPair struct {
}

func init() {
	i := AWSEc2KeyPair{}
	listers = append(listers, i)
}

func (l AWSEc2KeyPair) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2KeyPair}
}

func (l AWSEc2KeyPair) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	res, err := svc.DescribeKeyPairs(ctx.Context, &ec2.DescribeKeyPairsInput{})
	if err != nil {
		return rg, err
	}
	for _, kp := range res.KeyPairs {
		r := resource.New(ctx, resource.Ec2KeyPair, kp.KeyName, kp.KeyName, kp)
		rg.AddResource(r)
	}
	return rg, nil
}
