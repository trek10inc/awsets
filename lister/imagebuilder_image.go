package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSImageBuilderImage struct {
}

func init() {
	i := AWSImageBuilderImage{}
	listers = append(listers, i)
}

func (l AWSImageBuilderImage) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ImageBuilderImage,
	}
}

func (l AWSImageBuilderImage) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := imagebuilder.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListImages(ctx.Context, &imagebuilder.ListImagesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list imagebuilder images: %w", err)
		}
		for _, ivl := range res.ImageVersionList {
			v, err := svc.GetImage(ctx.Context, &imagebuilder.GetImageInput{
				ImageBuildVersionArn: ivl.Arn,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get imagebuilder image %s: %w", *ivl.Name, err)
			}
			imArn := arn.ParseP(v.Image.Arn)
			r := resource.New(ctx, resource.ImageBuilderImage, imArn.ResourceId, v.Image.Name, v.Image)
			r.AddARNRelation(resource.ImageBuilderImagePipeline, v.Image.SourcePipelineArn)

			// This response includes the full recipe, but only building the relationship to the ARN, then querying
			// the recipes separately to pick up any recipes that don't have images (yet?)
			r.AddARNRelation(resource.ImageBuilderImageRecipe, v.Image.ImageRecipe.Arn)

			// This response includes the full infrastructure config, but only building the relationship to the ARN,
			// then querying the configurations separately to pick up any configs that don't have images (yet?)
			r.AddARNRelation(resource.ImageBuilderInfrastructureConfiguration, v.Image.InfrastructureConfiguration.Arn)

			// This response includes the full distribution config, but only building the relationship to the ARN,
			// then querying the configurations separately to pick up any configs that don't have images (yet?)
			r.AddARNRelation(resource.ImageBuilderDistributionConfiguration, v.Image.DistributionConfiguration.Arn)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
