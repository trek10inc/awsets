package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := ec2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeNetworkInterfaces(ctx.Context, &ec2.DescribeNetworkInterfacesInput{
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, eni := range res.NetworkInterfaces {
			r := resource.New(ctx, resource.Ec2NetworkInterface, eni.NetworkInterfaceId, eni.NetworkInterfaceId, eni)
			r.AddRelation(resource.Ec2Vpc, eni.VpcId, "")
			r.AddRelation(resource.Ec2Subnet, eni.SubnetId, "")
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
