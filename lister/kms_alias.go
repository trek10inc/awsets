package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSKmsAlias struct {
}

func init() {
	i := AWSKmsAlias{}
	listers = append(listers, i)
}

func (l AWSKmsAlias) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.KmsAlias}
}

func (l AWSKmsAlias) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := kms.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListAliases(ctx.Context, &kms.ListAliasesInput{
			Limit:  aws.Int32(100),
			Marker: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, alias := range res.Aliases {
			aliasArn := arn.ParseP(alias.AliasArn)
			r := resource.New(ctx, resource.KmsAlias, aliasArn.ResourceId, alias.AliasName, alias)
			if alias.TargetKeyId != nil {
				r.AddRelation(resource.KmsKey, alias.TargetKeyId, "")
			}
			rg.AddResource(r)
		}
		return res.NextMarker, nil
	})
	return rg, err
}
