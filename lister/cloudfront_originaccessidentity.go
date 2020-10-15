package lister

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

var listCloudfrontOAIOnce sync.Once

type AWSCloudfrontOriginAccessIdentify struct {
}

func init() {
	i := AWSCloudfrontOriginAccessIdentify{}
	listers = append(listers, i)
}

func (l AWSCloudfrontOriginAccessIdentify) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CloudFrontOriginAccessIdentity}
}

func (l AWSCloudfrontOriginAccessIdentify) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := cloudfront.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listCloudfrontOAIOnce.Do(func() {

		err := Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListCloudFrontOriginAccessIdentities(cfg.Context, &cloudfront.ListCloudFrontOriginAccessIdentitiesInput{
				MaxItems: aws.String("100"),
				Marker:   nt,
			})
			if err != nil {
				return nil, err
			}
			if res.CloudFrontOriginAccessIdentityList == nil {
				return nil, nil
			}
			for _, item := range res.CloudFrontOriginAccessIdentityList.Items {
				r := resource.NewGlobal(cfg, resource.CloudFrontOriginAccessIdentity, item.Id, item.Id, item)
				rg.AddResource(r)
			}
			return res.CloudFrontOriginAccessIdentityList.NextMarker, nil
		})
		if err != nil {
			outerErr = err
		}
	})

	return rg, outerErr
}
