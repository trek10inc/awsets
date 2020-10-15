package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

var listCloudfrontCachePolicyOnce sync.Once

type AWSCloudfrontCachePolicy struct {
}

func init() {
	i := AWSCloudfrontCachePolicy{}
	listers = append(listers, i)
}

func (l AWSCloudfrontCachePolicy) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CloudFrontCachePolicy}
}

func (l AWSCloudfrontCachePolicy) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := cloudfront.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error
	listCloudfrontCachePolicyOnce.Do(func() {
		err := Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListCachePolicies(cfg.Context, &cloudfront.ListCachePoliciesInput{
				Marker:   nt,
				MaxItems: aws.String("100"),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list cache policies: %w", err)
			}
			if policies := res.CachePolicyList; policies != nil {
				for _, v := range policies.Items {
					r := resource.NewGlobal(cfg, resource.CloudFrontCachePolicy, v.CachePolicy.Id, v.CachePolicy.Id, v)
					rg.AddResource(r)
				}
				return policies.NextMarker, nil
			} else {
				return nil, nil
			}
		})
		if err != nil {
			outerErr = err
		}
	})

	return rg, outerErr
}
