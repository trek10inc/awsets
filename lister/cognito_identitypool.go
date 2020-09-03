package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
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
		resource.CognitoUserPool,
		resource.CognitoUserPoolClient,
		resource.CognitoUserPoolGroup,
		resource.CognitoUserPoolIdentityProvider,
		resource.CognitoUserPoolResourceServer,
	}
}

func (l AWSCognitoIdentityPool) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := cognitoidentity.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string

	for {
		identityPools, err := svc.ListIdentityPoolsRequest(&cognitoidentity.ListIdentityPoolsInput{
			MaxResults: aws.Int64(60),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list identity pools: %w", err)
		}

		for _, identityPool := range identityPools.IdentityPools {
			pool, err := svc.DescribeIdentityPoolRequest(&cognitoidentity.DescribeIdentityPoolInput{
				IdentityPoolId: identityPool.IdentityPoolId,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to describe identity pool %s: %w", *identityPool.IdentityPoolName, err)
			}
			r := resource.New(ctx, resource.CognitoIdentityPool, pool.IdentityPoolId, pool.IdentityPoolName, pool)
			rg.AddResource(r)
		}

		if identityPools.NextToken == nil {
			break
		}
		nextToken = identityPools.NextToken
	}

	return rg, nil
}
