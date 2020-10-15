package lister

import (
	"github.com/trek10inc/awsets/option"
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

func (l AWSEc2Vpc) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ec2.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeVpcs(cfg.Context, &ec2.DescribeVpcsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.Vpcs {
			r := resource.New(cfg, resource.Ec2Vpc, v.VpcId, v.VpcId, v)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
