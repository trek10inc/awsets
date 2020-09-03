package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/resource"
)

type AWSApiGatewayRestApi struct {
}

func init() {
	i := AWSApiGatewayRestApi{}
	listers = append(listers, i)
}

func (l AWSApiGatewayRestApi) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ApiGatewayRestApi,
		resource.ApiGatewayModel,
		resource.ApiGatewayDeployment,
		resource.ApiGatewayStage,
		resource.ApiGatewayAuthorizer,
		resource.ApiGatewayResource,
	}
}

func (l AWSApiGatewayRestApi) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := apigateway.New(ctx.AWSCfg)

	req := svc.GetRestApisRequest(&apigateway.GetRestApisInput{
		Limit: aws.Int64(500),
	})

	rg := resource.NewGroup()
	paginator := apigateway.NewGetRestApisPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, restapi := range page.Items {
			r := resource.New(ctx, resource.ApiGatewayRestApi, restapi.Id, restapi.Name, restapi)

			modelPaginator := apigateway.NewGetModelsPaginator(svc.GetModelsRequest(&apigateway.GetModelsInput{
				Limit:     aws.Int64(100),
				RestApiId: restapi.Id,
			}))
			for modelPaginator.Next(ctx.Context) {
				modelPage := modelPaginator.CurrentPage()
				for _, model := range modelPage.Items {
					modelR := resource.New(ctx, resource.ApiGatewayModel, model.Id, model.Name, model)
					modelR.AddRelation(resource.ApiGatewayRestApi, restapi.Id, "")
					rg.AddResource(modelR)
				}
			}
			if err := modelPaginator.Err(); err != nil {
				return rg, fmt.Errorf("failed to get models for api %s: %w", *restapi.Id, err)
			}

			depReq := svc.GetDeploymentsRequest(&apigateway.GetDeploymentsInput{
				Limit:     aws.Int64(500),
				RestApiId: restapi.Id,
			})
			depPaginator := apigateway.NewGetDeploymentsPaginator(depReq)
			for depPaginator.Next(ctx.Context) {
				depPage := depPaginator.CurrentPage()
				for _, deployment := range depPage.Items {
					depR := resource.New(ctx, resource.ApiGatewayDeployment, deployment.Id, "", restapi)
					r.AddRelation(resource.ApiGatewayDeployment, deployment.Id, "")
					rg.AddResource(depR)

					stageRes, err := svc.GetStagesRequest(&apigateway.GetStagesInput{
						DeploymentId: deployment.Id,
						RestApiId:    restapi.Id,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to get stages for api: %s, deployment: %s - %w", aws.StringValue(restapi.Id), aws.StringValue(deployment.Id), err)
					}
					for _, stage := range stageRes.Item {
						stageR := resource.New(ctx, resource.ApiGatewayStage, stage.StageName, "", stage)
						stageR.AddRelation(resource.ApiGatewayDeployment, stage.DeploymentId, "")
						stageR.AddRelation(resource.ApiGatewayRestApi, restapi.Id, "")
						if arn.IsArnP(stage.WebAclArn) {
							webAclArn := arn.ParseP(stage.WebAclArn)
							stageR.AddRelation(resource.WafRegionalWebACL, webAclArn.ResourceId, webAclArn.ResourceVersion)
						}
						rg.AddResource(stageR)
					}
				}
			}
			if err := depPaginator.Err(); err != nil {
				return rg, fmt.Errorf("failed to get deployments for restapi %s: %w", *restapi.Id, err)
			}

			var position *string
			for {
				authorizers, err := svc.GetAuthorizersRequest(&apigateway.GetAuthorizersInput{
					Limit:     aws.Int64(100),
					Position:  position,
					RestApiId: restapi.Id,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to get authorizers for rest api %s: %w", *restapi.Id, err)
				}
				for _, authorizer := range authorizers.Items {
					authR := resource.New(ctx, resource.ApiGatewayAuthorizer, authorizer.Id, authorizer.Name, authorizer)
					authR.AddRelation(resource.ApiGatewayRestApi, restapi.Id, "")
					rg.AddResource(authR)
				}
				if authorizers.Position == nil {
					break
				}
				position = authorizers.Position
			}

			resourcePaginator := apigateway.NewGetResourcesPaginator(svc.GetResourcesRequest(&apigateway.GetResourcesInput{
				Limit:     aws.Int64(100),
				RestApiId: restapi.Id,
			}))
			for resourcePaginator.Next(ctx.Context) {
				resourcesPage := resourcePaginator.CurrentPage()
				for _, res := range resourcesPage.Items {
					resR := resource.New(ctx, resource.ApiGatewayResource, res.Id, res.Id, res)
					resR.AddRelation(resource.ApiGatewayRestApi, restapi.Id, "")
					rg.AddResource(resR)
				}
			}
			if err := resourcePaginator.Err(); err != nil {
				return rg, fmt.Errorf("failed to get resources for restapi %s: %w", *restapi.Id, err)
			}

			var gwPosition *string
			gwResponses := make([]apigateway.GatewayResponse, 0)
			for {
				gwRes, err := svc.GetGatewayResponsesRequest(&apigateway.GetGatewayResponsesInput{
					Limit:     aws.Int64(100),
					Position:  gwPosition,
					RestApiId: restapi.Id,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to get gateway responses for rest api %s: %w", *restapi.Id, err)
				}
				if len(gwRes.Items) > 0 {
					gwResponses = append(gwResponses, gwRes.Items...)
				}
				if gwRes.Position == nil {
					break
				}
				gwPosition = gwRes.Position
			}
			if len(gwResponses) > 0 {
				r.AddAttribute("GatewayResponse", gwResponses)
			}
			rg.AddResource(r)

		}
	}
	err := paginator.Err()
	return rg, err
}
