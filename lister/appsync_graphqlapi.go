package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/resource"
)

type AWSAppsyncGraphqlApi struct {
}

func init() {
	i := AWSAppsyncGraphqlApi{}
	listers = append(listers, i)
}

func (l AWSAppsyncGraphqlApi) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.AppSyncGraphQLApi,
		resource.AppSyncApiKey,
		resource.AppSyncDataSource,
	}
}

func (l AWSAppsyncGraphqlApi) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := appsync.New(ctx.AWSCfg)

	rg := resource.NewGroup()

	var nextToken *string
	for {
		apis, err := svc.ListGraphqlApisRequest(&appsync.ListGraphqlApisInput{
			MaxResults: aws.Int64(25),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list graphql apis: %w", err)
		}
		for _, api := range apis.GraphqlApis {
			r := resource.New(ctx, resource.AppSyncGraphQLApi, api.ApiId, api.Name, api)
			//TODO: relation to user pool?
			rg.AddResource(r)

			var dsToken *string
			for {
				datasources, err := svc.ListDataSourcesRequest(&appsync.ListDataSourcesInput{
					ApiId:      api.ApiId,
					MaxResults: aws.Int64(25),
					NextToken:  dsToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list datasources for api %s: %w", aws.StringValue(api.ApiId), err)
				}
				for _, ds := range datasources.DataSources {
					dsArn := arn.ParseP(ds.DataSourceArn)
					dsRes := resource.New(ctx, resource.AppSyncDataSource, dsArn.ResourceId, "", ds)
					dsRes.AddRelation(resource.AppSyncGraphQLApi, api.ApiId, "")
					if ds.ServiceRoleArn != nil {
						srArn := arn.ParseP(ds.ServiceRoleArn)
						dsRes.AddRelation(resource.IamRole, srArn.ResourceId, "")
					}
					rg.AddResource(dsRes)
				}
				if datasources.NextToken == nil {
					break
				}
				dsToken = datasources.NextToken
			}

			var funcToken *string
			for {
				functions, err := svc.ListFunctionsRequest(&appsync.ListFunctionsInput{
					ApiId:      api.ApiId,
					MaxResults: aws.Int64(25),
					NextToken:  funcToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to get functions for %s: %w", *api.ApiId, err)
				}
				for _, function := range functions.Functions {
					funcR := resource.New(ctx, resource.AppSyncFunction, function.FunctionId, function.Name, function)
					funcR.AddRelation(resource.AppSyncGraphQLApi, api.ApiId, "")

					var resolverToken *string
					for {
						resolvers, err := svc.ListResolversByFunctionRequest(&appsync.ListResolversByFunctionInput{
							ApiId:      api.ApiId,
							FunctionId: function.FunctionId,
							MaxResults: aws.Int64(25),
							NextToken:  resolverToken,
						}).Send(ctx.Context)
						if err != nil {
							return rg, fmt.Errorf("failed to list resolvers for api %s: %w", aws.StringValue(api.ApiId), err)
						}
						for _, resolver := range resolvers.Resolvers {
							resolverArn := arn.ParseP(resolver.ResolverArn)
							resolverR := resource.New(ctx, resource.AppSyncResolver, resolverArn.ResourceId, resolverArn.ResourceId, resolver)
							resolverR.AddRelation(resource.AppSyncGraphQLApi, api.ApiId, "")
							resolverR.AddRelation(resource.AppSyncFunction, function.FunctionId, "")
							rg.AddResource(resolverR)
						}
						if resolvers.NextToken == nil {
							break
						}
						resolverToken = resolvers.NextToken
					}

					rg.AddResource(funcR)
				}
				if functions.NextToken == nil {
					break
				}
				funcToken = functions.NextToken
			}

			var apiKeyToken *string
			for {
				apiKeys, err := svc.ListApiKeysRequest(&appsync.ListApiKeysInput{
					ApiId:      api.ApiId,
					MaxResults: aws.Int64(25),
					NextToken:  apiKeyToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to get api keys for %s: %w", *api.ApiId, err)
				}
				for _, apiKey := range apiKeys.ApiKeys {
					keyR := resource.New(ctx, resource.AppSyncApiKey, apiKey.Id, apiKey.Id, apiKey)
					keyR.AddRelation(resource.AppSyncGraphQLApi, api.ApiId, "")
					rg.AddResource(keyR)
				}
				if apiKeys.NextToken == nil {
					break
				}
				apiKeyToken = apiKeys.NextToken
			}
		}
		if apis.NextToken == nil {
			break
		}
		nextToken = apis.NextToken
	}

	return rg, nil
}
