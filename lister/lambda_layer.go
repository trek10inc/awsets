package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/trek10inc/awsets/arn"
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
	svc := lambda.New(ctx.AWSCfg)

	req := svc.ListLayersRequest(&lambda.ListLayersInput{
		MaxItems: aws.Int64(50),
	})

	rg := resource.NewGroup()
	paginator := lambda.NewListLayersPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, layer := range page.Layers {
			layerArn := arn.ParseP(layer.LayerArn)

			r := resource.New(ctx, resource.LambdaLayer, layerArn.ResourceId, layer.LayerName, layer)
			rg.AddResource(r)

			layerReq := svc.ListLayerVersionsRequest(&lambda.ListLayerVersionsInput{
				LayerName: layer.LayerArn,
				MaxItems:  aws.Int64(50),
			})
			layerRes, err := layerReq.Send(ctx.Context)
			if err != nil {
				return rg, err
			}
			for _, lv := range layerRes.LayerVersions {
				lvArn := arn.ParseP(lv.LayerVersionArn)
				lvr := resource.New(ctx, resource.LambdaLayerVersion, lvArn.ResourceId, lvArn.ResourceVersion, lv)
				lvr.AddRelation(resource.LambdaLayer, layerArn.ResourceId, "")
				rg.AddResource(lvr)
			}
		}
	}
	err := paginator.Err()
	return rg, err
}
