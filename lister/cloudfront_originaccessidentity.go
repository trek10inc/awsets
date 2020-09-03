package lister

import (
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
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
	svc := cloudfront.New(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listCloudfrontOAIOnce.Do(func() {
		req := svc.ListCloudFrontOriginAccessIdentitiesRequest(&cloudfront.ListCloudFrontOriginAccessIdentitiesInput{
			MaxItems: aws.Int64(100),
		})

		paginator := cloudfront.NewListCloudFrontOriginAccessIdentitiesPaginator(req)
		for paginator.Next(ctx.Context) {
			page := paginator.CurrentPage()
			if page.CloudFrontOriginAccessIdentityList == nil {
				continue
			}
			for _, item := range page.CloudFrontOriginAccessIdentityList.Items {
				r := resource.NewGlobal(ctx, resource.CloudFrontOriginAccessIdentity, item.Id, item.Id, item)
				rg.AddResource(r)
			}
		}
		outerErr = paginator.Err()
	})

	return rg, outerErr
}
