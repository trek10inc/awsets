package lister

import (
	"strings"
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/trek10inc/awsets/resource"
)

var listCloudfrontDistributionsOnce sync.Once

type AWSCloudfrontDistribution struct {
}

func init() {
	i := AWSCloudfrontDistribution{}
	listers = append(listers, i)
}

func (l AWSCloudfrontDistribution) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CloudFrontDistribution}
}

func (l AWSCloudfrontDistribution) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloudfront.New(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listCloudfrontDistributionsOnce.Do(func() {
		req := svc.ListDistributionsRequest(&cloudfront.ListDistributionsInput{
			MaxItems: aws.Int64(100),
		})

		paginator := cloudfront.NewListDistributionsPaginator(req)
		for paginator.Next(ctx.Context) {
			page := paginator.CurrentPage()
			if page.DistributionList == nil {
				continue
			}
			for _, item := range page.DistributionList.Items {
				r := resource.NewGlobal(ctx, resource.CloudFrontDistribution, item.Id, item.Id, item)
				if item.Origins != nil {
					for _, origin := range item.Origins.Items {
						if origin.S3OriginConfig != nil && origin.S3OriginConfig.OriginAccessIdentity != nil {
							oai := strings.TrimPrefix(*origin.S3OriginConfig.OriginAccessIdentity, "origin-access-identity/cloudfront/")
							r.AddRelation(resource.CloudFrontOriginAccessIdentity, oai, "")
						}
					}
				}
				rg.AddResource(r)
			}
		}
		outerErr = paginator.Err()
	})

	return rg, outerErr
}
