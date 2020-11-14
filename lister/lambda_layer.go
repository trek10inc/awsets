package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSLambdaLayerVersion struct {
}

func init() {
	i := AWSLambdaLayerVersion{}
	listers = append(listers, i)
}

func (l AWSLambdaLayerVersion) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.LambdaLayer, resource.LambdaLayerVersion}
}

func (l AWSLambdaLayerVersion) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := lambda.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListLayers(ctx.Context, &lambda.ListLayersInput{
			MaxItems: aws.Int32(50),
			Marker:   nt,
		})
		if err != nil {
			return nil, err
		}
		for _, layer := range res.Layers {
			layerArn := arn.ParseP(layer.LayerArn)

			r := resource.New(ctx, resource.LambdaLayer, layerArn.ResourceId, layer.LayerName, layer)
			rg.AddResource(r)

			// Layer Versions
			err = Paginator(func(nt2 *string) (*string, error) {
				versions, err := svc.ListLayerVersions(ctx.Context, &lambda.ListLayerVersionsInput{
					LayerName: layer.LayerArn,
					MaxItems:  aws.Int32(50),
					Marker:    nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get versions for layer %s: %w", *layer.LayerName, err)
				}
				for _, lv := range versions.LayerVersions {
					lvArn := arn.ParseP(lv.LayerVersionArn)
					lvr := resource.New(ctx, resource.LambdaLayerVersion, lvArn.ResourceId, lvArn.ResourceVersion, lv)
					lvr.AddRelation(resource.LambdaLayer, layerArn.ResourceId, "")
					rg.AddResource(lvr)
				}
				return versions.NextMarker, nil
			})
			if err != nil {
				return nil, err
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
