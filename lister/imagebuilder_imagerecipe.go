package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSImageBuilderImageRecipe struct {
}

func init() {
	i := AWSImageBuilderImageRecipe{}
	listers = append(listers, i)
}

func (l AWSImageBuilderImageRecipe) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ImageBuilderImageRecipe,
	}
}

func (l AWSImageBuilderImageRecipe) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := imagebuilder.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListImageRecipes(cfg.Context, &imagebuilder.ListImageRecipesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list imagebuilder image recipes: %w", err)
		}
		for _, recipeSummary := range res.ImageRecipeSummaryList {
			v, err := svc.GetImageRecipe(cfg.Context, &imagebuilder.GetImageRecipeInput{
				ImageRecipeArn: recipeSummary.Arn,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get imagebuilder image recipe %s: %w", *recipeSummary.Name, err)
			}
			recipeArn := arn.ParseP(v.ImageRecipe.Arn)
			r := resource.New(cfg, resource.ImageBuilderImageRecipe, recipeArn.ResourceId, v.ImageRecipe.Name, v.ImageRecipe)
			if v.ImageRecipe.BlockDeviceMappings != nil {
				for _, bdm := range v.ImageRecipe.BlockDeviceMappings {
					if bdm.Ebs != nil {
						r.AddARNRelation(resource.KmsKey, bdm.Ebs.KmsKeyId)
						r.AddRelation(resource.Ec2Snapshot, bdm.Ebs.SnapshotId, "")
					}
				}
			}
			for _, c := range v.ImageRecipe.Components {
				r.AddARNRelation(resource.ImageBuilderComponent, c.ComponentArn)
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
