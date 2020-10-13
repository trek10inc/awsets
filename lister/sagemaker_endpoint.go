package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSSagemakerEndpoint struct {
}

func init() {
	i := AWSSagemakerEndpoint{}
	listers = append(listers, i)
}

func (l AWSSagemakerEndpoint) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.SagemakerEndpoint,
	}
}

func (l AWSSagemakerEndpoint) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := sagemaker.NewFromConfig(ctx.AWSCfg)

	res, err := svc.ListEndpoints(ctx.Context, &sagemaker.ListEndpointsInput{
		MaxResults: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := sagemaker.NewListEndpointsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, ep := range page.Endpoints {
			epRes, err := svc.DescribeEndpoint(ctx.Context, &sagemaker.DescribeEndpointInput{
				EndpointName: ep.EndpointName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe endpoint %s: %w", *ep.EndpointName, err)
			}
			v := epRes.DescribeEndpointOutput
			if v == nil {
				continue
			}
			epArn := arn.ParseP(v.EndpointArn)
			r := resource.New(ctx, resource.SagemakerEndpoint, epArn.ResourceId, v.EndpointName, v)
			if v.DataCaptureConfig != nil {
				r.AddARNRelation(resource.KmsKey, v.DataCaptureConfig.KmsKeyId)
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
