package lister

import (
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/trek10inc/awsets/resource"
)

var listCloudfrontStreamingDistributionsOnce sync.Once

type AWSCloudfrontStreamingDistribution struct {
}

func init() {
	i := AWSCloudfrontStreamingDistribution{}
	listers = append(listers, i)
}

func (l AWSCloudfrontStreamingDistribution) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CloudFrontStreamingDistribution}
}

func (l AWSCloudfrontStreamingDistribution) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloudfront.New(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listCloudfrontStreamingDistributionsOnce.Do(func() {
		req := svc.ListStreamingDistributionsRequest(&cloudfront.ListStreamingDistributionsInput{
			MaxItems: aws.Int64(100),
		})

		paginator := cloudfront.NewListStreamingDistributionsPaginator(req)
		for paginator.Next(ctx.Context) {
			page := paginator.CurrentPage()
			if page.StreamingDistributionList == nil {
				continue
			}
			for _, item := range page.StreamingDistributionList.Items {
				r := resource.NewGlobal(ctx, resource.CloudFrontStreamingDistribution, item.Id, item.Id, item)
				if item.S3Origin != nil {
					r.AddRelation(resource.CloudFrontOriginAccessIdentity, item.S3Origin.OriginAccessIdentity, "")
				}
				rg.AddResource(r)
			}
		}
		outerErr = paginator.Err()
	})

	return rg, outerErr
}
