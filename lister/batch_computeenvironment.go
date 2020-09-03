package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/batch"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSBatchComputeEnvironment struct {
}

func init() {
	i := AWSBatchComputeEnvironment{}
	listers = append(listers, i)
}

func (l AWSBatchComputeEnvironment) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.BatchComputeEnvironment}
}

func (l AWSBatchComputeEnvironment) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := batch.New(ctx.AWSCfg)

	req := svc.DescribeComputeEnvironmentsRequest(&batch.DescribeComputeEnvironmentsInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := batch.NewDescribeComputeEnvironmentsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.ComputeEnvironments {
			r := resource.New(ctx, resource.BatchComputeEnvironment, v.ComputeEnvironmentName, v.ComputeEnvironmentName, v)
			if c := v.ComputeResources; c != nil {
				r.AddRelation(resource.Ec2Image, c.ImageId, "")
				r.AddRelation(resource.Ec2KeyPair, c.Ec2KeyPair, "")
				for _, sn := range c.Subnets {
					r.AddRelation(resource.Ec2Subnet, sn, "")
				}
				for _, sg := range c.SecurityGroupIds {
					r.AddRelation(resource.Ec2SecurityGroup, sg, "")
				}
				r.AddARNRelation(resource.IamRole, c.InstanceRole)
				r.AddARNRelation(resource.IamRole, c.SpotIamFleetRole)
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
