package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSImageBuilderImagePipeline struct {
}

func init() {
	i := AWSImageBuilderImagePipeline{}
	listers = append(listers, i)
}

func (l AWSImageBuilderImagePipeline) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ImageBuilderImagePipeline,
	}
}

func (l AWSImageBuilderImagePipeline) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := imagebuilder.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListImagePipelines(ctx.Context, &imagebuilder.ListImagePipelinesInput{
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list imagebuilder image pipelines: %w", err)
		}
		for _, v := range res.ImagePipelineList {
			plArn := arn.ParseP(v.Arn)
			r := resource.New(ctx, resource.ImageBuilderImagePipeline, plArn.ResourceId, v.Name, v)
			r.AddARNRelation(resource.ImageBuilderImageRecipe, v.ImageRecipeArn)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
