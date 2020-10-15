package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/trek10inc/awsets/option"
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

func (l AWSApiGatewayV2Api) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := apigatewayv2.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.GetApis(cfg.Context, &apigatewayv2.GetApisInput{
			MaxResults: aws.String("100"),
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list apigatewayv2 apis: %w", err)
		}
		for _, v := range res.Items {
			r := resource.New(cfg, resource.ApiGatewayV2Api, v.ApiId, v.Name, v)

			// Authorizers
			err = Paginator(func(nt2 *string) (*string, error) {
				authRes, err := svc.GetAuthorizers(cfg.Context, &apigatewayv2.GetAuthorizersInput{
					ApiId:      v.ApiId,
					MaxResults: aws.String("100"),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list apigatewayv2 api authorizers for api %s: %w", *v.ApiId, err)
				}
				for _, authorizer := range authRes.Items {
					authR := resource.New(cfg, resource.ApiGatewayV2Authorizer, authorizer.AuthorizerId, authorizer.Name, authorizer)
					authR.AddRelation(resource.ApiGatewayV2Api, v.ApiId, v.Version)
					rg.AddResource(authR)
				}
				return authRes.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Deployments
			err = Paginator(func(nt2 *string) (*string, error) {
				deploymentsRes, err := svc.GetDeployments(cfg.Context, &apigatewayv2.GetDeploymentsInput{
					ApiId:      v.ApiId,
					MaxResults: aws.String("100"),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list apigatewayv2 api deployments for api %s: %w", *v.ApiId, err)
				}
				for _, deployment := range deploymentsRes.Items {
					deployR := resource.New(cfg, resource.ApiGatewayV2Deployment, deployment.DeploymentId, deployment.DeploymentId, deployment)
					deployR.AddRelation(resource.ApiGatewayV2Api, v.ApiId, v.Version)
					rg.AddResource(deployR)
				}
				return deploymentsRes.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Integrations
			err = Paginator(func(nt2 *string) (*string, error) {
				intRes, err := svc.GetIntegrations(cfg.Context, &apigatewayv2.GetIntegrationsInput{
					ApiId:      v.ApiId,
					MaxResults: aws.String("100"),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list apigatewayv2 integrations for api %s: %w", *v.ApiId, err)
				}
				for _, integration := range intRes.Items {
					intR := resource.New(cfg, resource.ApiGatewayV2Integration, integration.IntegrationId, integration.IntegrationId, integration)
					intR.AddRelation(resource.ApiGatewayV2Api, v.ApiId, v.Version)
					rg.AddResource(intR)
				}
				return intRes.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Routes
			err = Paginator(func(nt2 *string) (*string, error) {
				routesRes, err := svc.GetRoutes(cfg.Context, &apigatewayv2.GetRoutesInput{
					ApiId:      v.ApiId,
					MaxResults: aws.String("100"),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list apigatewayv2 routes for api %s: %w", *v.ApiId, err)
				}
				for _, route := range routesRes.Items {
					routeR := resource.New(cfg, resource.ApiGatewayV2Integration, route.RouteId, route.RouteId, route)
					routeR.AddRelation(resource.ApiGatewayAuthorizer, route.AuthorizerId, "")
					routeR.AddRelation(resource.ApiGatewayV2Api, v.ApiId, v.Version)
					rg.AddResource(routeR)
				}
				return routesRes.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Stages
			err = Paginator(func(nt2 *string) (*string, error) {
				stagesRes, err := svc.GetStages(cfg.Context, &apigatewayv2.GetStagesInput{
					ApiId:      v.ApiId,
					MaxResults: aws.String("100"),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list apigatewayv2 stages for api %s: %w", *v.ApiId, err)
				}
				for _, stage := range stagesRes.Items {
					stageR := resource.New(cfg, resource.ApiGatewayV2Stage, stage.StageName, stage.StageName, stage)
					stageR.AddRelation(resource.ApiGatewayV2Deployment, stage.DeploymentId, "")
					stageR.AddRelation(resource.ApiGatewayV2Api, v.ApiId, v.Version)
					rg.AddResource(stageR)
				}
				return stagesRes.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Models
			err = Paginator(func(nt2 *string) (*string, error) {
				modelsRes, err := svc.GetModels(cfg.Context, &apigatewayv2.GetModelsInput{
					ApiId:      v.ApiId,
					MaxResults: aws.String("100"),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list apigatewayv2 models for api %s: %w", *v.ApiId, err)
				}
				for _, model := range modelsRes.Items {
					modelR := resource.New(cfg, resource.ApiGatewayV2Model, model.ModelId, model.Name, model)
					modelR.AddRelation(resource.ApiGatewayV2Api, v.ApiId, v.Version)
					rg.AddResource(modelR)
				}
				return modelsRes.NextToken, nil
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
