package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSSagemakerModel struct {
}

func init() {
	i := AWSSagemakerModel{}
	listers = append(listers, i)
}

func (l AWSSagemakerModel) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.SagemakerModel}
}

func (l AWSSagemakerModel) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := sagemaker.NewFromConfig(ctx.AWSCfg)

	res, err := svc.ListModels(ctx.Context, &sagemaker.ListModelsInput{
		MaxResults: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := sagemaker.NewListModelsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, model := range page.Models {
			modelRes, err := svc.DescribeModel(ctx.Context, &sagemaker.DescribeModelInput{
				ModelName: model.ModelName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe model %s: %w", *model.ModelName, err)
			}
			v := modelRes.DescribeModelOutput
			if v == nil {
				continue
			}
			modelArn := arn.ParseP(v.ModelArn)
			r := resource.New(ctx, resource.SagemakerModel, modelArn.ResourceId, v.ModelName, v)
			r.AddARNRelation(resource.IamRole, v.ExecutionRoleArn)
			if vpc := v.VpcConfig; vpc != nil {
				for _, sg := range vpc.SecurityGroupIds {
					r.AddRelation(resource.Ec2SecurityGroup, sg, "")
				}
				for _, sn := range vpc.Subnets {
					r.AddRelation(resource.Ec2Subnet, sn, "")
				}
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
