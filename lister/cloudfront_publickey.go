package lister

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

var listCloudfrontPublicKeyOnce sync.Once

type AWSCloudfrontPublicKey struct {
}

func init() {
	i := AWSCloudfrontPublicKey{}
	listers = append(listers, i)
}

func (l AWSCloudfrontPublicKey) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CloudFrontPublicKey}
}

func (l AWSCloudfrontPublicKey) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloudfront.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listCloudfrontPublicKeyOnce.Do(func() {
		err := Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListPublicKeys(ctx.Context, &cloudfront.ListPublicKeysInput{
				MaxItems: aws.String("100"),
				Marker:   nt,
			})
			if err != nil {
				return nil, err
			}
			if res.PublicKeyList == nil {
				return nil, nil
			}
			for _, item := range res.PublicKeyList.Items {
				r := resource.NewGlobal(ctx, resource.CloudFrontPublicKey, item.Id, item.Name, item)

				rg.AddResource(r)
			}
			return res.PublicKeyList.NextMarker, nil
		})
		if err != nil {
			outerErr = err
		}
	})

	return rg, outerErr
}
