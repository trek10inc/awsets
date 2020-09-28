package lister

import (
	"fmt"
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/trek10inc/awsets/resource"
)

var listCloudfrontOriginRequestPolicyOnce sync.Once

type AWSCloudfrontOriginRequestPolicy struct {
}

func init() {
	i := AWSCloudfrontOriginRequestPolicy{}
	listers = append(listers, i)
}

func (l AWSCloudfrontOriginRequestPolicy) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CloudFrontOriginRequestPolicy}
}

func (l AWSCloudfrontOriginRequestPolicy) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloudfront.New(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error
	listCloudfrontOriginRequestPolicyOnce.Do(func() {

		var marker *string
		for {
			res, err := svc.ListOriginRequestPoliciesRequest(&cloudfront.ListOriginRequestPoliciesInput{
				Marker:   marker,
				MaxItems: aws.Int64(100),
			}).Send(ctx.Context)
			if err != nil {
				outerErr = fmt.Errorf("failed to list origin request policies: %w", err)
				return
			}
			if policies := res.OriginRequestPolicyList; policies != nil {
				for _, v := range policies.Items {
					r := resource.NewGlobal(ctx, resource.CloudFrontOriginRequestPolicy, v.OriginRequestPolicy.Id, v.OriginRequestPolicy.Id, v)
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
