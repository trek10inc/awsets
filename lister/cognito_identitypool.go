package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSCognitoIdentityPool struct {
}

func init() {
	i := AWSCognitoIdentityPool{}
	listers = append(listers, i)
}

func (l AWSCognitoIdentityPool) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.CognitoIdentityPool,
	}
}

func (l AWSCognitoIdentityPool) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := cognitoidentity.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListIdentityPools(cfg.Context, &cognitoidentity.ListIdentityPoolsInput{
			MaxResults: aws.Int32(60),
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list identity pools: %w", err)
		}
		for _, identityPool := range res.IdentityPools {
			pool, err := svc.DescribeIdentityPool(cfg.Context, &cognitoidentity.DescribeIdentityPoolInput{
				IdentityPoolId: identityPool.IdentityPoolId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe identity pool %s: %w", *identityPool.IdentityPoolName, err)
			}
			r := resource.New(cfg, resource.CognitoIdentityPool, pool.IdentityPoolId, pool.IdentityPoolName, pool)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
