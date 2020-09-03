package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/arn"
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
	svc := cognitoidentityprovider.New(ctx.AWSCfg)

	req := svc.ListUserPoolsRequest(&cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int64(60),
	})

	rg := resource.NewGroup()
	paginator := cognitoidentityprovider.NewListUserPoolsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.UserPools {
			userPoolResponse, err := svc.DescribeUserPoolRequest(&cognitoidentityprovider.DescribeUserPoolInput{
				UserPoolId: v.Id,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get user pool %s: %w", *v.Id, err)
			}
			up := userPoolResponse.UserPool
			r := resource.New(ctx, resource.CognitoUserPool, up.Id, up.Name, up)
			for tagName, tagValue := range up.UserPoolTags {
				r.Tags[tagName] = tagValue
			}

			clientPaginator := cognitoidentityprovider.NewListUserPoolClientsPaginator(svc.ListUserPoolClientsRequest(&cognitoidentityprovider.ListUserPoolClientsInput{
				UserPoolId: v.Id,
			}))
			for clientPaginator.Next(ctx.Context) {
				clientPage := clientPaginator.CurrentPage()
				for _, client := range clientPage.UserPoolClients {

					clientResponse, err := svc.DescribeUserPoolClientRequest(&cognitoidentityprovider.DescribeUserPoolClientInput{
						ClientId:   client.ClientId,
						UserPoolId: client.UserPoolId,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to get client %s for pool %s: %w", *client.ClientId, *v.Id, err)
					}
					c := clientResponse.UserPoolClient
					clientR := resource.New(ctx, resource.CognitoUserPoolClient, c.ClientId, c.ClientName, c)
					clientR.AddRelation(resource.CognitoUserPool, client.UserPoolId, "")
					r.AddRelation(resource.CognitoUserPoolClient, c.ClientId, "")
					rg.AddResource(clientR)
				}
			}
			err = clientPaginator.Err()
			if err != nil {
				return rg, fmt.Errorf("failed to list clients for user pool %s: %w", *up.Id, err)
			}

			groupPaginator := cognitoidentityprovider.NewListGroupsPaginator(svc.ListGroupsRequest(&cognitoidentityprovider.ListGroupsInput{
				Limit:      aws.Int64(60),
				UserPoolId: v.Id,
			}))
			for groupPaginator.Next(ctx.Context) {
				groupPage := groupPaginator.CurrentPage()
				for _, group := range groupPage.Groups {
					groupR := resource.New(ctx, resource.CognitoUserPoolGroup, group.GroupName, group.GroupName, group)
					groupR.AddRelation(resource.CognitoUserPool, group.UserPoolId, "")
					if group.RoleArn != nil {
						roleArn := arn.ParseP(group.RoleArn)
						groupR.AddRelation(resource.IamRole, roleArn.ResourceId, roleArn.ResourceVersion)
					}
					rg.AddResource(groupR)
				}
			}
			err = groupPaginator.Err()
			if err != nil {
				return rg, fmt.Errorf("failed to list groups for user pool %s: %w", *up.Id, err)
			}

			identifyProviderPaginator := cognitoidentityprovider.NewListIdentityProvidersPaginator(svc.ListIdentityProvidersRequest(&cognitoidentityprovider.ListIdentityProvidersInput{
				MaxResults: aws.Int64(60),
				UserPoolId: up.Id,
			}))
			for identifyProviderPaginator.Next(ctx.Context) {
				ipPage := identifyProviderPaginator.CurrentPage()
				for _, ip := range ipPage.Providers {
					ipR := resource.New(ctx, resource.CognitoUserPoolGroup, ip.ProviderName, ip.ProviderName, ip)
					ipR.AddRelation(resource.CognitoUserPool, up.Id, "")
					rg.AddResource(ipR)
					r.AddRelation(resource.CognitoUserPoolIdentityProvider, ip.ProviderName, "")
				}
			}
			err = identifyProviderPaginator.Err()
			if err != nil {
				return rg, fmt.Errorf("failed to list identity providers for user pool %s: %w", *up.Id, err)
			}

			rsPaginator := cognitoidentityprovider.NewListResourceServersPaginator(svc.ListResourceServersRequest(&cognitoidentityprovider.ListResourceServersInput{
				MaxResults: aws.Int64(50),
				UserPoolId: up.Id,
			}))
			for rsPaginator.Next(ctx.Context) {
				rsPage := rsPaginator.CurrentPage()
				for _, rs := range rsPage.ResourceServers {
					rsR := resource.New(ctx, resource.CognitoUserPoolResourceServer, rs.Identifier, rs.Name, rs)
					rsR.AddRelation(resource.CognitoUserPool, up.Id, "")
					rg.AddResource(rsR)
					r.AddRelation(resource.CognitoUserPoolResourceServer, rs.Identifier, "")
				}
			}
			err = rsPaginator.Err()
			if err != nil {
				return rg, fmt.Errorf("failed to list resource servers for user pool %s: %w", *up.Id, err)
			}

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
