package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
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
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeKeyPairsRequest(&ec2.DescribeKeyPairsInput{})

	rg := resource.NewGroup()
	res, err := req.Send(ctx.Context)
	if err != nil {
		return rg, err
	}
	for _, kp := range res.KeyPairs {
		r := resource.New(ctx, resource.Ec2KeyPair, kp.KeyName, kp.KeyName, kp)
		rg.AddResource(r)
	}
	return rg, nil
}
