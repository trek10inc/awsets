package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/trek10inc/awsets/resource"
)

type AWSApiGatewayApiKey struct {
}

func init() {
	i := AWSApiGatewayApiKey{}
	listers = append(listers, i)
}

func (l AWSApiGatewayApiKey) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ApiGatewayApiKey, resource.ApiGatewayUsagePlan, resource.ApiGatewayUsagePlanKey}
}

func (l AWSApiGatewayApiKey) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := apigateway.New(ctx.AWSCfg)

	req := svc.GetApiKeysRequest(&apigateway.GetApiKeysInput{
		Limit: aws.Int64(500),
	})

	rg := resource.NewGroup()
	paginator := apigateway.NewGetApiKeysPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, apikey := range page.Items {
			r := resource.New(ctx, resource.ApiGatewayRestApi, apikey.Id, apikey.Name, apikey)
			rg.AddResource(r)
			usagePaginator := apigateway.NewGetUsagePlansPaginator(svc.GetUsagePlansRequest(&apigateway.GetUsagePlansInput{
				KeyId: apikey.Id,
				Limit: aws.Int64(500),
			}))
			for usagePaginator.Next(ctx.Context) {
				usagePlanPage := usagePaginator.CurrentPage()
				for _, usagePlan := range usagePlanPage.Items {
					usagePlanRes := resource.New(ctx, resource.ApiGatewayUsagePlan, usagePlan.Id, usagePlan.Name, usagePlan)
					usagePlanRes.AddRelation(resource.ApiGatewayApiKey, apikey.Id, "")
					for _, stage := range usagePlan.ApiStages {
						usagePlanRes.AddRelation(resource.ApiGatewayStage, stage.Stage, "")
					}
					rg.AddResource(usagePlanRes)

					usageKeyPaginator := apigateway.NewGetUsagePlanKeysPaginator(svc.GetUsagePlanKeysRequest(&apigateway.GetUsagePlanKeysInput{
						Limit:       aws.Int64(10),
						UsagePlanId: usagePlan.Id,
					}))
					for usageKeyPaginator.Next(ctx.Context) {
						usagePlanKeyPage := usageKeyPaginator.CurrentPage()
						for _, usagePlanKey := range usagePlanKeyPage.Items {
							planKeyRes := resource.New(ctx, resource.ApiGatewayUsagePlanKey, usagePlanKey.Id, usagePlanKey.Name, usagePlanKey)
							planKeyRes.AddRelation(resource.ApiGatewayUsagePlan, usagePlan.Id, "")
							rg.AddResource(planKeyRes)
						}
					}
					if err := usageKeyPaginator.Err(); err != nil {
						return rg, fmt.Errorf("failed to get usage plan keys for %s: %w", aws.StringValue(usagePlan.Id), err)
					}
				}
			}
			if err := usagePaginator.Err(); err != nil {
				return rg, fmt.Errorf("failed to get usage plans for %s: %w", aws.StringValue(apikey.Id), err)
			}
		}
	}
	err := paginator.Err()
	return rg, err
}
