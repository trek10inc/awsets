package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSSagemakerNotebookInstanceLifecycleConfig struct {
}

func init() {
	i := AWSSagemakerNotebookInstanceLifecycleConfig{}
	listers = append(listers, i)
}

func (l AWSSagemakerNotebookInstanceLifecycleConfig) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.SagemakerNotebookInstanceLifecycleConfig,
	}
}

func (l AWSSagemakerNotebookInstanceLifecycleConfig) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := sagemaker.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListNotebookInstanceLifecycleConfigs(ctx.Context, &sagemaker.ListNotebookInstanceLifecycleConfigsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, config := range res.NotebookInstanceLifecycleConfigs {
			v, err := svc.DescribeNotebookInstanceLifecycleConfig(ctx.Context, &sagemaker.DescribeNotebookInstanceLifecycleConfigInput{
				NotebookInstanceLifecycleConfigName: config.NotebookInstanceLifecycleConfigName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe sagemaker notebook instance lifecycle name %s: %w", *config.NotebookInstanceLifecycleConfigName, err)
			}
			configArn := arn.ParseP(v.NotebookInstanceLifecycleConfigArn)
			r := resource.New(ctx, resource.SagemakerNotebookInstanceLifecycleConfig, configArn.ResourceId, v.NotebookInstanceLifecycleConfigName, v)

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
