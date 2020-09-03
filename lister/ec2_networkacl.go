package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
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
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeNetworkAclsRequest(&ec2.DescribeNetworkAclsInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := ec2.NewDescribeNetworkAclsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, nacl := range page.NetworkAcls {
			r := resource.New(ctx, resource.Ec2NetworkACL, nacl.NetworkAclId, nacl.NetworkAclId, nacl)
			r.AddRelation(resource.Ec2Vpc, nacl.VpcId, "")
			for _, a := range nacl.Associations {
				r.AddRelation(resource.Ec2Subnet, a.SubnetId, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
