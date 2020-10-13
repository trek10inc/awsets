package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/trek10inc/awsets/arn"
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

	res, err := svc.ListAliases(ctx.Context, &kms.ListAliasesInput{
		Limit: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := kms.NewListAliasesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, alias := range page.Aliases {
			aliasArn := arn.ParseP(alias.AliasArn)
			r := resource.New(ctx, resource.KmsAlias, aliasArn.ResourceId, alias.AliasName, alias)
			if alias.TargetKeyId != nil {
				r.AddRelation(resource.KmsKey, alias.TargetKeyId, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
