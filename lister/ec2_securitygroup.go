package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2SecurityGroup struct {
}

func init() {
	i := AWSEc2SecurityGroup{}
	listers = append(listers, i)
}

func (l AWSEc2SecurityGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2SecurityGroup}
}

func (l AWSEc2SecurityGroup) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ec2.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeSecurityGroups(cfg.Context, &ec2.DescribeSecurityGroupsInput{
			MaxResults: aws.Int32(1000),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.SecurityGroups {
			r := resource.New(cfg, resource.Ec2SecurityGroup, v.GroupId, v.GroupName, v)
			if v.VpcId != nil {
				r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
