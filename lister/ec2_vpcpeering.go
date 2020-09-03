package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AWSEc2VpcPeering struct {
}

func init() {
	i := AWSEc2VpcPeering{}
	listers = append(listers, i)
}

func (l AWSEc2VpcPeering) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2VpcPeering}
}

func (l AWSEc2VpcPeering) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeVpcPeeringConnectionsRequest(&ec2.DescribeVpcPeeringConnectionsInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := ec2.NewDescribeVpcPeeringConnectionsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.VpcPeeringConnections {
			r := resource.New(ctx, resource.Ec2VpcPeering, v.VpcPeeringConnectionId, v.VpcPeeringConnectionId, v)
			if v.AccepterVpcInfo != nil {
				r.AddCrossRelation(ctx.AccountId, v.AccepterVpcInfo.Region, resource.Ec2Vpc, v.AccepterVpcInfo.VpcId, "")
			}
			if v.RequesterVpcInfo != nil {
				r.AddCrossRelation(ctx.AccountId, v.RequesterVpcInfo.Region, resource.Ec2Vpc, v.RequesterVpcInfo.VpcId, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
