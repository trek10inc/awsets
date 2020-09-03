package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSApiGatewayV2Api struct {
}

func init() {
	i := AWSApiGatewayV2Api{}
	listers = append(listers, i)
}

func (l AWSApiGatewayV2Api) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ApiGatewayV2Api,
		resource.ApiGatewayV2Authorizer,
		resource.ApiGatewayV2Deployment,
		resource.ApiGatewayV2Integration,
	}
}

func (l AWSApiGatewayV2Api) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := apigatewayv2.New(ctx.AWSCfg)

	rg := resource.NewGroup()

	var nextToken *string
	for {
		res, err := svc.GetApisRequest(&apigatewayv2.GetApisInput{
			MaxResults: aws.String("100"),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list apigatewayv2 apis: %w", err)
		}
		if res.GetApisOutput == nil {
			continue
		}
		for _, v := range res.GetApisOutput.Items {
			r := resource.New(ctx, resource.ApiGatewayV2Api, v.ApiId, v.Name, v)

			// Authorizers
			var authNextToken *string
			for {
				authRes, err := svc.GetAuthorizersRequest(&apigatewayv2.GetAuthorizersInput{
					ApiId:      v.ApiId,
					MaxResults: aws.String("100"),
					NextToken:  authNextToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list apigatewayv2 api authorizers for api %s: %w", *v.ApiId, err)
				}
				for _, authorizer := range authRes.Items {
					authR := resource.New(ctx, resource.ApiGatewayV2Authorizer, authorizer.AuthorizerId, authorizer.Name, authorizer)
					authR.AddRelation(resource.ApiGatewayV2Api, v.ApiId, v.Version)
					rg.AddResource(authR)
				}
				if authRes.NextToken == nil {
					break
				}
				authNextToken = authRes.NextToken
			}

			// Deployments
			var deployNextToken *string
			for {
				deploymentsRes, err := svc.GetDeploymentsRequest(&apigatewayv2.GetDeploymentsInput{
					ApiId:      v.ApiId,
					MaxResults: aws.String("100"),
					NextToken:  deployNextToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list apigatewayv2 api deployments for api %s: %w", *v.ApiId, err)
				}
				for _, deployment := range deploymentsRes.Items {
					deployR := resource.New(ctx, resource.ApiGatewayV2Deployment, deployment.DeploymentId, deployment.DeploymentId, deployment)
					deployR.AddRelation(resource.ApiGatewayV2Api, v.ApiId, v.Version)
					rg.AddResource(deployR)
				}
				if deploymentsRes.NextToken == nil {
					break
				}
				deployNextToken = deploymentsRes.NextToken
			}

			// Integrations
			var integrationToken *string
			for {
				intRes, err := svc.GetIntegrationsRequest(&apigatewayv2.GetIntegrationsInput{
					ApiId:      v.ApiId,
					MaxResults: aws.String("100"),
					NextToken:  integrationToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list apigatewayv2 integrations for api %s: %w", *v.ApiId, err)
				}
				for _, integration := range intRes.Items {
					intR := resource.New(ctx, resource.ApiGatewayV2Integration, integration.IntegrationId, integration.IntegrationId, integration)
					intR.AddRelation(resource.ApiGatewayV2Api, v.ApiId, v.Version)
					rg.AddResource(intR)
				}
				if intRes.NextToken == nil {
					break
				}
				deployNextToken = intRes.NextToken
			}

			// Routes
			var routesToken *string
			for {
				routesRes, err := svc.GetRoutesRequest(&apigatewayv2.GetRoutesInput{
					ApiId:      v.ApiId,
					MaxResults: aws.String("100"),
					NextToken:  routesToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list apigatewayv2 routes for api %s: %w", *v.ApiId, err)
				}
				for _, route := range routesRes.Items {
					routeR := resource.New(ctx, resource.ApiGatewayV2Integration, route.RouteId, route.RouteId, route)
					routeR.AddRelation(resource.ApiGatewayAuthorizer, route.AuthorizerId, "")
					routeR.AddRelation(resource.ApiGatewayV2Api, v.ApiId, v.Version)
					rg.AddResource(routeR)
				}
				if routesRes.NextToken == nil {
					break
				}
				routesToken = routesRes.NextToken
			}

			// Stages
			var stagesToken *string
			for {
				stagesRes, err := svc.GetStagesRequest(&apigatewayv2.GetStagesInput{
					ApiId:      v.ApiId,
					MaxResults: aws.String("100"),
					NextToken:  stagesToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list apigatewayv2 stages for api %s: %w", *v.ApiId, err)
				}
				for _, stage := range stagesRes.Items {
					stageR := resource.New(ctx, resource.ApiGatewayV2Stage, stage.StageName, stage.StageName, stage)
					stageR.AddRelation(resource.ApiGatewayV2Deployment, stage.DeploymentId, "")
					stageR.AddRelation(resource.ApiGatewayV2Api, v.ApiId, v.Version)
					rg.AddResource(stageR)
				}
				if stagesRes.NextToken == nil {
					break
				}
				stagesToken = stagesRes.NextToken
			}

			// Models
			var modelsToken *string
			for {
				modelsRes, err := svc.GetModelsRequest(&apigatewayv2.GetModelsInput{
					ApiId:      v.ApiId,
					MaxResults: aws.String("100"),
					NextToken:  modelsToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list apigatewayv2 models for api %s: %w", *v.ApiId, err)
				}
				for _, model := range modelsRes.Items {
					modelR := resource.New(ctx, resource.ApiGatewayV2Model, model.ModelId, model.Name, model)
					modelR.AddRelation(resource.ApiGatewayV2Api, v.ApiId, v.Version)
					rg.AddResource(modelR)
				}
				if modelsRes.NextToken == nil {
					break
				}
				modelsToken = modelsRes.NextToken
			}

			rg.AddResource(r)
		}
		if res.NextToken == nil {
			break
		}
		nextToken = res.NextToken
	}
	return rg, nil
}
