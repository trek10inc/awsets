package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AWSEc2NetworkInterface struct {
}

func init() {
	i := AWSEc2NetworkInterface{}
	listers = append(listers, i)
}

func (l AWSEc2NetworkInterface) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2NetworkInterface}
}

func (l AWSEc2NetworkInterface) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeNetworkInterfacesRequest(&ec2.DescribeNetworkInterfacesInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := ec2.NewDescribeNetworkInterfacesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, eni := range page.NetworkInterfaces {
			r := resource.New(ctx, resource.Ec2NetworkInterface, eni.NetworkInterfaceId, eni.NetworkInterfaceId, eni)
			r.AddRelation(resource.Ec2Vpc, eni.VpcId, "")
			r.AddRelation(resource.Ec2Subnet, eni.SubnetId, "")
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
