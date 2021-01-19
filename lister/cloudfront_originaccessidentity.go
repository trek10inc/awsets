package lister

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/trek10inc/awsets/context"
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

func (l AWSCloudfrontOriginAccessIdentify) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloudfront.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listCloudfrontOAIOnce.Do(func() {

		err := Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListCloudFrontOriginAccessIdentities(ctx.Context, &cloudfront.ListCloudFrontOriginAccessIdentitiesInput{
				MaxItems: aws.Int32(100),
				Marker:   nt,
			})
			if err != nil {
				return nil, err
			}
			if res.CloudFrontOriginAccessIdentityList == nil {
				return nil, nil
			}
			for _, item := range res.CloudFrontOriginAccessIdentityList.Items {
				r := resource.NewGlobal(ctx, resource.CloudFrontOriginAccessIdentity, item.Id, item.Id, item)
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
