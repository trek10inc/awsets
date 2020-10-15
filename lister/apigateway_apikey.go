package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSApiGatewayApiKey struct {
}

func init() {
	i := AWSApiGatewayApiKey{}
	listers = append(listers, i)
}

func (l AWSApiGatewayApiKey) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ApiGatewayApiKey,
		resource.ApiGatewayUsagePlan,
		resource.ApiGatewayUsagePlanKey,
	}
}

func (l AWSApiGatewayApiKey) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := apigateway.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		req, err := svc.GetApiKeys(cfg.Context, &apigateway.GetApiKeysInput{
			Position: nt,
			Limit:    aws.Int32(500),
		})
		if err != nil {
			if strings.Contains(err.Error(), "AccessDeniedException") {
				err = nil
			}
			return nil, err
		}
		for _, apikey := range req.Items {
			r := resource.New(cfg, resource.ApiGatewayRestApi, apikey.Id, apikey.Name, apikey)
			rg.AddResource(r)

			// Usage Plans
			err = Paginator(func(nt2 *string) (*string, error) {
				usagePlanReq, err := svc.GetUsagePlans(cfg.Context, &apigateway.GetUsagePlansInput{
					KeyId: apikey.Id,
					Limit: aws.Int32(500),
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get usage plans for %s: %w", *apikey.Id, err)
				}
				for _, usagePlan := range usagePlanReq.Items {
					usagePlanRes := resource.New(cfg, resource.ApiGatewayUsagePlan, usagePlan.Id, usagePlan.Name, usagePlan)
					usagePlanRes.AddRelation(resource.ApiGatewayApiKey, apikey.Id, "")
					for _, stage := range usagePlan.ApiStages {
						usagePlanRes.AddRelation(resource.ApiGatewayStage, stage.Stage, "")
					}
					rg.AddResource(usagePlanRes)

					// Usage Plan Keys
					err = Paginator(func(nt3 *string) (*string, error) {
						planKeysRes, err := svc.GetUsagePlanKeys(cfg.Context, &apigateway.GetUsagePlanKeysInput{
							Limit:       aws.Int32(10),
							Position:    nt3,
							UsagePlanId: usagePlan.Id,
						})
						if err != nil {
							return nil, fmt.Errorf("failed to get usage plan keys for plan %s: %w", *usagePlan.Id, err)
						}
						for _, usagePlanKey := range planKeysRes.Items {
							planKeyRes := resource.New(cfg, resource.ApiGatewayUsagePlanKey, usagePlanKey.Id, usagePlanKey.Name, usagePlanKey)
							planKeyRes.AddRelation(resource.ApiGatewayUsagePlan, usagePlan.Id, "")
							rg.AddResource(planKeyRes)
						}
						return planKeysRes.Position, nil
					})
				}
				return usagePlanReq.Position, nil
			})
			if err != nil {
				return nil, err
			}
		}
		return req.Position, nil
	})
	return rg, err
}
