package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSKmsKey struct {
}

func init() {
	i := AWSKmsKey{}
	listers = append(listers, i)
}

func (l AWSKmsKey) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.KmsKey}
}

func (l AWSKmsKey) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := kms.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListKeys(ctx.Context, &kms.ListKeysInput{
			Limit:  aws.Int32(100),
			Marker: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, key := range res.Keys {
			keyDetail, err := svc.DescribeKey(ctx.Context, &kms.DescribeKeyInput{
				GrantTokens: nil,
				KeyId:       key.KeyId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe key %s: %w", *key.KeyId, err)
			}
			if v := keyDetail.KeyMetadata; v != nil {
				r := resource.New(ctx, resource.KmsKey, v.KeyId, v.KeyId, v)
				// TODO: relationshio to HSM?
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
