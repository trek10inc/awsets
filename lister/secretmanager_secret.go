package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
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

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListSecrets(ctx.Context, &secretsmanager.ListSecretsInput{
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.SecretList {
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
		return res.NextToken, nil
	})
	return rg, err
}
