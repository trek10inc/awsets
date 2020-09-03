package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
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

func (l AWSEc2SecurityGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeSecurityGroupsRequest(&ec2.DescribeSecurityGroupsInput{
		MaxResults: aws.Int64(1000),
	})

	rg := resource.NewGroup()
	paginator := ec2.NewDescribeSecurityGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.SecurityGroups {
			r := resource.New(ctx, resource.Ec2SecurityGroup, v.GroupId, v.GroupName, v)
			if v.VpcId != nil {
				r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
