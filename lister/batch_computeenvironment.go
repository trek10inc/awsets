package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/batch"
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
	svc := batch.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeComputeEnvironments(ctx.Context, &batch.DescribeComputeEnvironmentsInput{
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.ComputeEnvironments {
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

		return res.NextToken, nil
	})
	return rg, err
}
