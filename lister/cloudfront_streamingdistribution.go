package lister

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/trek10inc/awsets/context"
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
	svc := cloudfront.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listCloudfrontStreamingDistributionsOnce.Do(func() {
		outerErr = Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListStreamingDistributions(ctx.Context, &cloudfront.ListStreamingDistributionsInput{
				MaxItems: aws.Int32(100),
				Marker:   nt,
			})
			if err != nil {
				return nil, err
			}
			if res.StreamingDistributionList != nil {
				return nil, nil
			}
			for _, item := range res.StreamingDistributionList.Items {
				r := resource.NewGlobal(ctx, resource.CloudFrontStreamingDistribution, item.Id, item.Id, item)
				if item.S3Origin != nil {
					r.AddRelation(resource.CloudFrontOriginAccessIdentity, item.S3Origin.OriginAccessIdentity, "")
				}
				rg.AddResource(r)
			}
			return res.StreamingDistributionList.NextMarker, nil
		})
	})

	return rg, outerErr
}
