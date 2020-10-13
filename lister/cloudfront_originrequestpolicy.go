package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/trek10inc/awsets/context"
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
	svc := cloudfront.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error
	listCloudfrontOriginRequestPolicyOnce.Do(func() {
		err := Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListOriginRequestPolicies(ctx.Context, &cloudfront.ListOriginRequestPoliciesInput{
				Marker:   nt,
				MaxItems: aws.String("100"),
			})
			if err != nil {
				return nil, err
			}
			if policies := res.OriginRequestPolicyList; policies != nil {
				for _, v := range policies.Items {
					r := resource.NewGlobal(ctx, resource.CloudFrontOriginRequestPolicy, v.OriginRequestPolicy.Id, v.OriginRequestPolicy.Id, v)
					rg.AddResource(r)
				}
				return policies.NextMarker, nil
			} else {
				return nil, nil
			}
		})
		if err != nil {
			outerErr = fmt.Errorf("failed to list origin request policies: %w", err)
		}
	})

	return rg, outerErr
}
