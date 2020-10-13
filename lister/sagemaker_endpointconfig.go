package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSSagemakerEndpointConfig struct {
}

func init() {
	i := AWSSagemakerEndpointConfig{}
	listers = append(listers, i)
}

func (l AWSSagemakerEndpointConfig) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.SagemakerEndpointConfig,
	}
}

func (l AWSSagemakerEndpointConfig) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := sagemaker.NewFromConfig(ctx.AWSCfg)

	res, err := svc.ListEndpointConfigs(ctx.Context, &sagemaker.ListEndpointConfigsInput{
		MaxResults: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := sagemaker.NewListEndpointConfigsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, epc := range page.EndpointConfigs {
			epcRes, err := svc.DescribeEndpointConfig(ctx.Context, &sagemaker.DescribeEndpointConfigInput{
				EndpointConfigName: epc.EndpointConfigName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe endpoint config %s: %w", *epc.EndpointConfigName, err)
			}
			v := epcRes.DescribeEndpointConfigOutput
			if v == nil {
				continue
			}
			epcArn := arn.ParseP(v.EndpointConfigArn)
			r := resource.New(ctx, resource.SagemakerEndpointConfig, epcArn.ResourceId, v.EndpointConfigName, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			if v.DataCaptureConfig != nil {
				r.AddARNRelation(resource.KmsKey, v.DataCaptureConfig.KmsKeyId)
			}
			for _, pv := range v.ProductionVariants {
				r.AddRelation(resource.SagemakerModel, pv.ModelName, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
