package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AWSEc2Instance struct {
}

func init() {
	i := AWSEc2Instance{}
	listers = append(listers, i)
}

func (l AWSEc2Instance) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2Instance}
}

func (l AWSEc2Instance) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeInstancesRequest(&ec2.DescribeInstancesInput{
		MaxResults: aws.Int64(1000),
	})

	rg := resource.NewGroup()
	paginator := ec2.NewDescribeInstancesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, reservation := range page.Reservations {
			for _, v := range reservation.Instances {
				r := resource.New(ctx, resource.Ec2Instance, v.InstanceId, v.InstanceId, v)
				r.AddRelation(resource.Ec2Subnet, v.SubnetId, "")
				r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
				for _, sg := range v.SecurityGroups {
					r.AddRelation(resource.Ec2SecurityGroup, sg.GroupId, "")
				}
				for _, eni := range v.NetworkInterfaces {
					r.AddRelation(resource.Ec2NetworkInterface, eni.NetworkInterfaceId, "")
				}
				r.AddRelation(resource.Ec2KeyPair, v.KeyName, "")
				for _, bm := range v.BlockDeviceMappings {
					if bm.Ebs != nil {
						r.AddRelation(resource.Ec2Volume, bm.Ebs.VolumeId, "")
					}
				}
				rg.AddResource(r)
			}
		}
	}
	err := paginator.Err()
	return rg, err
}
