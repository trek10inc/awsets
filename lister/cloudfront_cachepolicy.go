package lister

import (
	"fmt"
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
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

func (l AWSCloudfrontCachePolicy) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloudfront.New(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error
	listCloudfrontCachePolicyOnce.Do(func() {

		var marker *string
		for {
			res, err := svc.ListCachePoliciesRequest(&cloudfront.ListCachePoliciesInput{
				Marker:   marker,
				MaxItems: aws.Int64(100),
			}).Send(ctx.Context)
			if err != nil {
				outerErr = fmt.Errorf("failed to list cache policies: %w", err)
				return
			}
			if policies := res.CachePolicyList; policies != nil {
				for _, v := range policies.Items {
					r := resource.NewGlobal(ctx, resource.CloudFrontCachePolicy, v.CachePolicy.Id, v.CachePolicy.Id, v)
					rg.AddResource(r)
				}
				if policies.NextMarker == nil {
					break
				}
				marker = policies.NextMarker
			} else {
				break
			}
		}
	})

	return rg, outerErr
}
