package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2NetworkACL struct {
}

func init() {
	i := AWSEc2NetworkACL{}
	listers = append(listers, i)
}

func (l AWSEc2NetworkACL) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2NetworkACL}
}

func (l AWSEc2NetworkACL) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeNetworkAcls(ctx.Context, &ec2.DescribeNetworkAclsInput{
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, nacl := range res.NetworkAcls {
			r := resource.New(ctx, resource.Ec2NetworkACL, nacl.NetworkAclId, nacl.NetworkAclId, nacl)
			r.AddRelation(resource.Ec2Vpc, nacl.VpcId, "")
			for _, a := range nacl.Associations {
				r.AddRelation(resource.Ec2Subnet, a.SubnetId, "")
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
