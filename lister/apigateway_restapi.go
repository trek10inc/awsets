package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
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
		resource.ApiGatewayMethod,
		resource.ApiGatewayRequestValidator,
		resource.ApiGatewayDocumentationPart,
	}
}

func (l AWSApiGatewayRestApi) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := apigateway.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.GetRestApis(ctx.Context, &apigateway.GetRestApisInput{
			Limit:    aws.Int32(500),
			Position: nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "AccessDeniedException") {
				// If api gateway is not supported in a region, returns access denied
				return nil, nil
			}
			return nil, fmt.Errorf("failed to get rest apis: %w", err)
		}
		for _, restapi := range res.Items {
			r := resource.New(ctx, resource.ApiGatewayRestApi, restapi.Id, restapi.Name, restapi)

			// Models
			err = Paginator(func(nt2 *string) (*string, error) {
				modelRes, err := svc.GetModels(ctx.Context, &apigateway.GetModelsInput{
					Limit:     aws.Int32(100),
					RestApiId: restapi.Id,
					Position:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get models for api %s: %w", *restapi.Id, err)
				}
				for _, model := range modelRes.Items {
					modelR := resource.New(ctx, resource.ApiGatewayModel, model.Id, model.Name, model)
					modelR.AddRelation(resource.ApiGatewayRestApi, restapi.Id, "")
					rg.AddResource(modelR)
				}
				return modelRes.Position, nil
			})
			if err != nil {
				return nil, err
			}

			// Deployments
			err = Paginator(func(nt2 *string) (*string, error) {
				depRes, err := svc.GetDeployments(ctx.Context, &apigateway.GetDeploymentsInput{
					Limit:     aws.Int32(500),
					RestApiId: restapi.Id,
					Position:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get deployments for restapi %s: %w", *restapi.Id, err)
				}
				for _, deployment := range depRes.Items {
					depR := resource.New(ctx, resource.ApiGatewayDeployment, deployment.Id, "", restapi)
					r.AddRelation(resource.ApiGatewayDeployment, deployment.Id, "")
					rg.AddResource(depR)

					stageRes, err := svc.GetStages(ctx.Context, &apigateway.GetStagesInput{
						DeploymentId: deployment.Id,
						RestApiId:    restapi.Id,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to get stages for api: %s, deployment: %s - %w", *restapi.Id, *deployment.Id, err)
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
				return depRes.Position, nil
			})
			if err != nil {
				return nil, err
			}

			// Authorizers
			err = Paginator(func(nt2 *string) (*string, error) {
				authorizers, err := svc.GetAuthorizers(ctx.Context, &apigateway.GetAuthorizersInput{
					Limit:     aws.Int32(100),
					Position:  nt2,
					RestApiId: restapi.Id,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get authorizers for rest api %s: %w", *restapi.Id, err)
				}
				for _, authorizer := range authorizers.Items {
					authR := resource.New(ctx, resource.ApiGatewayAuthorizer, authorizer.Id, authorizer.Name, authorizer)
					authR.AddRelation(resource.ApiGatewayRestApi, restapi.Id, "")
					rg.AddResource(authR)
				}
				return authorizers.Position, nil
			})
			if err != nil {
				return nil, err
			}

			// Resources
			err = Paginator(func(nt2 *string) (*string, error) {
				resourcesRes, err := svc.GetResources(ctx.Context, &apigateway.GetResourcesInput{
					Limit:     aws.Int32(100),
					RestApiId: restapi.Id,
					Position:  nt2,
					Embed:     []string{"methods"},
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get resources for restapi %s: %w", *restapi.Id, err)
				}
				for _, res := range resourcesRes.Items {
					resR := resource.New(ctx, resource.ApiGatewayResource, res.Id, res.Id, res)
					resR.AddRelation(resource.ApiGatewayRestApi, restapi.Id, "")
					rg.AddResource(resR)
					for verb, method := range res.ResourceMethods {
						methodId := fmt.Sprintf("%s-%s", *res.Id, verb)
						methodR := resource.New(ctx, resource.ApiGatewayMethod, methodId, methodId, method)
						methodR.AddRelation(resource.ApiGatewayResource, res.Id, "")
						rg.AddResource(methodR)
					}
				}
				return resourcesRes.Position, nil
			})
			if err != nil {
				return nil, err
			}

			// Gateway Responses
			gwResponses := make([]types.GatewayResponse, 0)
			err = Paginator(func(nt2 *string) (*string, error) {
				gwRes, err := svc.GetGatewayResponses(ctx.Context, &apigateway.GetGatewayResponsesInput{
					Limit:     aws.Int32(100),
					Position:  nt2,
					RestApiId: restapi.Id,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get gateway responses for rest api %s: %w", *restapi.Id, err)
				}
				if len(gwRes.Items) > 0 {
					gwResponses = append(gwResponses, gwRes.Items...)
				}
				return gwRes.Position, nil
			})
			if err != nil {
				return nil, err
			}
			if len(gwResponses) > 0 {
				r.AddAttribute("GatewayResponse", gwResponses)
			}

			// Request Validators
			err = Paginator(func(nt2 *string) (*string, error) {
				rvRes, err := svc.GetRequestValidators(ctx.Context, &apigateway.GetRequestValidatorsInput{
					RestApiId: restapi.Id,
					Limit:     aws.Int32(100),
					Position:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get request validators for restapi %s: %w", *restapi.Id, err)
				}
				for _, rv := range rvRes.Items {
					rvR := resource.New(ctx, resource.ApiGatewayRequestValidator, rv.Id, rv.Name, rv)
					rvR.AddRelation(resource.ApiGatewayRestApi, restapi.Id, "")
					rg.AddResource(rvR)
				}
				return rvRes.Position, nil
			})
			if err != nil {
				return nil, err
			}

			// Documentation Parts
			err = Paginator(func(nt2 *string) (*string, error) {
				dpRes, err := svc.GetDocumentationParts(ctx.Context, &apigateway.GetDocumentationPartsInput{
					RestApiId: restapi.Id,
					Limit:     aws.Int32(100),
					Position:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get documentation parts for restapi %s: %w", *restapi.Id, err)
				}
				for _, dp := range dpRes.Items {
					dpR := resource.New(ctx, resource.ApiGatewayDocumentationPart, dp.Id, dp.Id, dp)
					dpR.AddRelation(resource.ApiGatewayRestApi, restapi.Id, "")
					rg.AddResource(dpR)
				}
				return dpRes.Position, nil
			})
			if err != nil {
				return nil, err
			}

			rg.AddResource(r)

		}
		return res.Position, nil
	})
	return rg, err
}
