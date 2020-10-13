package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
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

	res, err := svc.ListKeys(ctx.Context, &kms.ListKeysInput{
		Limit: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := kms.NewListKeysPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, key := range page.Keys {
			kres, err := svc.DescribeKey(ctx.Context, &kms.DescribeKeyInput{
				GrantTokens: nil,
				KeyId:       key.KeyId,
			})
			kres, err := kreq
			if err != nil {
				return rg, err
			}
			if kres.KeyMetadata != nil {
				km := kres.KeyMetadata
				r := resource.New(ctx, resource.KmsKey, km.KeyId, km.Arn, km)
				rg.AddResource(r)
			}
		}
	}
	err := paginator.Err()
	return rg, err
}
