package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/appsync"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
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
		resource.AppSyncFunction,
		resource.AppSyncResolver,
		resource.AppSyncApiCache,
	}
}

func (l AWSAppsyncGraphqlApi) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := appsync.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		apis, err := svc.ListGraphqlApis(ctx.Context, &appsync.ListGraphqlApisInput{
			MaxResults: 25,
			NextToken:  nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "ForbiddenException") {
				// If appsync isn't supported in a region, it returns 403, ForbiddenException
				return nil, nil
			}
			return nil, fmt.Errorf("failed to list graphql apis: %w", err)
		}
		for _, api := range apis.GraphqlApis {
			r := resource.New(ctx, resource.AppSyncGraphQLApi, api.ApiId, api.Name, api)
			// TODO: relation to user pool?
			rg.AddResource(r)

			// DataSources
			err = Paginator(func(nt2 *string) (*string, error) {
				datasources, err := svc.ListDataSources(ctx.Context, &appsync.ListDataSourcesInput{
					ApiId:      api.ApiId,
					MaxResults: 25,
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list datasources for api %s: %w", *api.ApiId, err)
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
				return datasources.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Functions
			err = Paginator(func(nt2 *string) (*string, error) {
				functions, err := svc.ListFunctions(ctx.Context, &appsync.ListFunctionsInput{
					ApiId:      api.ApiId,
					MaxResults: 25,
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get functions for %s: %w", *api.ApiId, err)
				}
				for _, function := range functions.Functions {
					funcR := resource.New(ctx, resource.AppSyncFunction, function.FunctionId, function.Name, function)
					funcR.AddRelation(resource.AppSyncGraphQLApi, api.ApiId, "")

					err = Paginator(func(nt3 *string) (*string, error) {
						resolvers, err := svc.ListResolversByFunction(ctx.Context, &appsync.ListResolversByFunctionInput{
							ApiId:      api.ApiId,
							FunctionId: function.FunctionId,
							MaxResults: 25,
							NextToken:  nt3,
						})
						if err != nil {
							return nil, fmt.Errorf("failed to list resolvers for api %s: %w", *api.ApiId, err)
						}
						for _, resolver := range resolvers.Resolvers {
							resolverArn := arn.ParseP(resolver.ResolverArn)
							resolverR := resource.New(ctx, resource.AppSyncResolver, resolverArn.ResourceId, resolverArn.ResourceId, resolver)
							resolverR.AddRelation(resource.AppSyncGraphQLApi, api.ApiId, "")
							resolverR.AddRelation(resource.AppSyncFunction, function.FunctionId, "")
							rg.AddResource(resolverR)
						}
						return resolvers.NextToken, nil
					})
					if err != nil {
						return nil, err
					}

					rg.AddResource(funcR)
				}
				return functions.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// API Keys
			err = Paginator(func(nt2 *string) (*string, error) {
				apiKeys, err := svc.ListApiKeys(ctx.Context, &appsync.ListApiKeysInput{
					ApiId:      api.ApiId,
					MaxResults: 25,
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get api keys for %s: %w", *api.ApiId, err)
				}
				for _, apiKey := range apiKeys.ApiKeys {
					keyR := resource.New(ctx, resource.AppSyncApiKey, apiKey.Id, apiKey.Id, apiKey)
					keyR.AddRelation(resource.AppSyncGraphQLApi, api.ApiId, "")
					rg.AddResource(keyR)
				}
				return apiKeys.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Cache
			apiCache, err := svc.GetApiCache(ctx.Context, &appsync.GetApiCacheInput{
				ApiId: api.ApiId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get api cache for %s: %w", *api.ApiId, err)
			}
			if v := apiCache.ApiCache; v != nil {
				cacheR := resource.New(ctx, resource.AppSyncApiCache, fmt.Sprintf("%s-cache", *api.ApiId), fmt.Sprintf("%s-cache", *api.ApiId), v)
				cacheR.AddRelation(resource.AppSyncGraphQLApi, api.ApiId, "")
				rg.AddResource(cacheR)
			}
		}
		return apis.NextToken, nil
	})
	return rg, err
}
