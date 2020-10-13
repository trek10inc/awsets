package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSSecretManagerSecret struct {
}

func init() {
	i := AWSSecretManagerSecret{}
	listers = append(listers, i)
}

func (l AWSSecretManagerSecret) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.SecretManagerSecret}
}

func (l AWSSecretManagerSecret) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := secretsmanager.NewFromConfig(ctx.AWSCfg)

	res, err := svc.ListSecrets(ctx.Context, &secretsmanager.ListSecretsInput{
		MaxResults: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := secretsmanager.NewListSecretsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.SecretList {
			r := resource.New(ctx, resource.SecretManagerSecret, v.Name, v.Name, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			policy, err := svc.GetResourcePolicy(ctx.Context, &secretsmanager.GetResourcePolicyInput{
				SecretId: v.Name,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get secret policy for %s: %w", *v.Name, err)
			}
			r.AddAttribute("ResourcePolicy", policy.ResourcePolicy)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
