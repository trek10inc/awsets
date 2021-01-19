package lister

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

var listCloudfrontKeyGroupOnce sync.Once

type AWSCloudfrontKeyGroup struct {
}

func init() {
	i := AWSCloudfrontKeyGroup{}
	listers = append(listers, i)
}

func (l AWSCloudfrontKeyGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CloudFrontKeyGroup}
}

func (l AWSCloudfrontKeyGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloudfront.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listCloudfrontKeyGroupOnce.Do(func() {
		err := Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListKeyGroups(ctx.Context, &cloudfront.ListKeyGroupsInput{
				MaxItems: aws.Int32(100),
				Marker:   nt,
			})
			if err != nil {
				return nil, err
			}
			if res.KeyGroupList == nil {
				return nil, nil
			}
			for _, item := range res.KeyGroupList.Items {
				kg := item.KeyGroup
				r := resource.NewGlobal(ctx, resource.CloudFrontKeyGroup, kg.Id, kg.Id, kg)

				rg.AddResource(r)
			}
			return res.KeyGroupList.NextMarker, nil
		})
		if err != nil {
			outerErr = err
		}
	})

	return rg, outerErr
}
