package lister

import (
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/trek10inc/awsets/option"
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

func (l AWSCloudfrontDistribution) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := cloudfront.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listCloudfrontDistributionsOnce.Do(func() {
		err := Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListDistributions(cfg.Context, &cloudfront.ListDistributionsInput{
				MaxItems: aws.String("100"),
				Marker:   nt,
			})
			if err != nil {
				return nil, err
			}
			if res.DistributionList == nil {
				return nil, nil
			}
			for _, item := range res.DistributionList.Items {
				r := resource.NewGlobal(cfg, resource.CloudFrontDistribution, item.Id, item.Id, item)
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
			return res.DistributionList.NextMarker, nil
		})
		if err != nil {
			outerErr = err
		}
	})

	return rg, outerErr
}
