package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/option"
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

func (l AWSEc2KeyPair) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ec2.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	res, err := svc.DescribeKeyPairs(cfg.Context, &ec2.DescribeKeyPairsInput{})
	if err != nil {
		return rg, err
	}
	for _, kp := range res.KeyPairs {
		r := resource.New(cfg, resource.Ec2KeyPair, kp.KeyName, kp.KeyName, kp)
		rg.AddResource(r)
	}
	return rg, nil
}
