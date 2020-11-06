package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iotsitewise"
	"github.com/trek10inc/awsets/option"
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

func (l AWSIoTSiteWiseAsset) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := iotsitewise.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListAssetModels(cfg.Context, &iotsitewise.ListAssetModelsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.AssetModelSummaries {
			r := resource.New(cfg, resource.IoTSiteWiseAssetModel, v.Id, v.Name, v)

			err = Paginator(func(nt2 *string) (*string, error) {
				assets, err := svc.ListAssets(cfg.Context, &iotsitewise.ListAssetsInput{
					MaxResults: aws.Int32(100),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list assets for model %s: %v", *v.Id, err)
				}
				for _, asset := range assets.AssetSummaries {
					aR := resource.New(cfg, resource.IoTSiteWiseAsset, asset.Id, asset.Name, asset)
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
