package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSCognitoUserpool struct {
}

func init() {
	i := AWSCognitoUserpool{}
	listers = append(listers, i)
}

func (l AWSCognitoUserpool) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.CognitoUserPool,
		resource.CognitoUserPoolClient,
		resource.CognitoUserPoolGroup,
		resource.CognitoUserPoolIdentityProvider,
		resource.CognitoUserPoolResourceServer,
	}
}

func (l AWSCognitoUserpool) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cognitoidentityprovider.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListUserPools(ctx.Context, &cognitoidentityprovider.ListUserPoolsInput{
			MaxResults: aws.Int32(60),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.UserPools {
			userPoolResponse, err := svc.DescribeUserPool(ctx.Context, &cognitoidentityprovider.DescribeUserPoolInput{
				UserPoolId: v.Id,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get user pool %s: %w", *v.Id, err)
			}
			up := userPoolResponse.UserPool
			r := resource.New(ctx, resource.CognitoUserPool, up.Id, up.Name, up)
			for tagName, tagValue := range up.UserPoolTags {
				r.Tags[tagName] = aws.ToString(tagValue)
			}

			// Clients
			err = Paginator(func(nt2 *string) (*string, error) {
				clients, err := svc.ListUserPoolClients(ctx.Context, &cognitoidentityprovider.ListUserPoolClientsInput{
					UserPoolId: v.Id,
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list clients for user pool %s: %w", *up.Id, err)
				}
				for _, client := range clients.UserPoolClients {

					clientResponse, err := svc.DescribeUserPoolClient(ctx.Context, &cognitoidentityprovider.DescribeUserPoolClientInput{
						ClientId:   client.ClientId,
						UserPoolId: client.UserPoolId,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to get client %s for pool %s: %w", *client.ClientId, *v.Id, err)
					}
					c := clientResponse.UserPoolClient
					clientR := resource.New(ctx, resource.CognitoUserPoolClient, c.ClientId, c.ClientName, c)
					clientR.AddRelation(resource.CognitoUserPool, client.UserPoolId, "")
					r.AddRelation(resource.CognitoUserPoolClient, c.ClientId, "")
					rg.AddResource(clientR)
				}
				return clients.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Groups
			err = Paginator(func(nt2 *string) (*string, error) {
				groups, err := svc.ListGroups(ctx.Context, &cognitoidentityprovider.ListGroupsInput{
					Limit:      aws.Int32(60),
					UserPoolId: v.Id,
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list groups for user pool %s: %w", *up.Id, err)
				}
				for _, group := range groups.Groups {
					groupR := resource.New(ctx, resource.CognitoUserPoolGroup, group.GroupName, group.GroupName, group)
					groupR.AddRelation(resource.CognitoUserPool, group.UserPoolId, "")
					groupR.AddARNRelation(resource.IamRole, group.RoleArn)
					rg.AddResource(groupR)
				}
				return groups.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Identity Providers
			err = Paginator(func(nt2 *string) (*string, error) {
				identityProviders, err := svc.ListIdentityProviders(ctx.Context, &cognitoidentityprovider.ListIdentityProvidersInput{
					MaxResults: aws.Int32(60),
					UserPoolId: up.Id,
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list identity providers for user pool %s: %w", *up.Id, err)
				}
				for _, ip := range identityProviders.Providers {
					ipR := resource.New(ctx, resource.CognitoUserPoolGroup, ip.ProviderName, ip.ProviderName, ip)
					ipR.AddRelation(resource.CognitoUserPool, up.Id, "")
					rg.AddResource(ipR)
					r.AddRelation(resource.CognitoUserPoolIdentityProvider, ip.ProviderName, "")
				}

				return identityProviders.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Resource Servers
			err = Paginator(func(nt2 *string) (*string, error) {
				resourceProviders, err := svc.ListResourceServers(ctx.Context, &cognitoidentityprovider.ListResourceServersInput{
					MaxResults: aws.Int32(50),
					UserPoolId: up.Id,
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list resource servers for user pool %s: %w", *up.Id, err)
				}
				for _, rs := range resourceProviders.ResourceServers {
					rsR := resource.New(ctx, resource.CognitoUserPoolResourceServer, rs.Identifier, rs.Name, rs)
					rsR.AddRelation(resource.CognitoUserPool, up.Id, "")
					rg.AddResource(rsR)
					r.AddRelation(resource.CognitoUserPoolResourceServer, rs.Identifier, "")
				}
				return resourceProviders.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
