package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/trek10inc/awsets/option"
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

func (l AWSSecretManagerSecret) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := secretsmanager.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListSecrets(cfg.Context, &secretsmanager.ListSecretsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.SecretList {
			r := resource.New(cfg, resource.SecretManagerSecret, v.Name, v.Name, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			policy, err := svc.GetResourcePolicy(cfg.Context, &secretsmanager.GetResourcePolicyInput{
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
