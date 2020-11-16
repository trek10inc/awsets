package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iotsitewise"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSIoTSiteWiseAsset struct {
}

func init() {
	i := AWSIoTSiteWiseAsset{}
	listers = append(listers, i)
}

func (l AWSIoTSiteWiseAsset) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.IoTSiteWiseAssetModel,
		resource.IoTSiteWiseAsset,
	}
}

func (l AWSIoTSiteWiseAsset) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := iotsitewise.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListAssetModels(ctx.Context, &iotsitewise.ListAssetModelsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.AssetModelSummaries {
			r := resource.New(ctx, resource.IoTSiteWiseAssetModel, v.Id, v.Name, v)

			err = Paginator(func(nt2 *string) (*string, error) {
				assets, err := svc.ListAssets(ctx.Context, &iotsitewise.ListAssetsInput{
					MaxResults: aws.Int32(100),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list assets for model %s: %v", *v.Id, err)
				}
				for _, asset := range assets.AssetSummaries {
					aR := resource.New(ctx, resource.IoTSiteWiseAsset, asset.Id, asset.Name, asset)
					aR.AddRelation(resource.IoTSiteWiseAssetModel, v.Id, "")
					rg.AddResource(aR)
				}
				return assets.NextToken, nil
			})

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
