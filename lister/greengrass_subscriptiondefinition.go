package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"

	"github.com/aws/aws-sdk-go-v2/service/greengrass"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSGreengrassSubscriptionDefinition struct {
}

func init() {
	i := AWSGreengrassSubscriptionDefinition{}
	listers = append(listers, i)
}

func (l AWSGreengrassSubscriptionDefinition) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GreengrassSubscriptionDefinition,
		resource.GreengrassSubscriptionDefinitionVersion,
	}
}

func (l AWSGreengrassSubscriptionDefinition) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := greengrass.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		subscriptiondefs, err := svc.ListSubscriptionDefinitionsRequest(&greengrass.ListSubscriptionDefinitionsInput{
			MaxResults: aws.String("100"),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				if aerr.Code() == "TooManyRequestsException" &&
					strings.Contains(aerr.Message(), "exceeded maximum number of attempts") {
					// If greengrass is not supported in a region, returns "TooManyRequests exception"
					return rg, nil
				}
			}
			return rg, fmt.Errorf("failed to list greengrass subscription definitions: %w", err)
		}
		for _, v := range subscriptiondefs.Definitions {
			r := resource.New(ctx, resource.GreengrassGroup, v.Id, v.Name, v)
			var sdNextToken *string
			for {
				subscriptionDefVersions, err := svc.ListSubscriptionDefinitionVersionsRequest(&greengrass.ListSubscriptionDefinitionVersionsInput{
					SubscriptionDefinitionId: v.Id,
					MaxResults:               aws.String("100"),
					NextToken:                sdNextToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list greengrass subscription definition versions for %s: %w", *v.Id, err)
				}
				for _, sdId := range subscriptionDefVersions.Versions {
					sd, err := svc.GetSubscriptionDefinitionVersionRequest(&greengrass.GetSubscriptionDefinitionVersionInput{
						SubscriptionDefinitionId:        sdId.Id,
						SubscriptionDefinitionVersionId: sdId.Version,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to list greengrass subscription definition version for %s, %s: %w", *sdId.Id, *sdId.Version, err)
					}
					sdRes := resource.NewVersion(ctx, resource.GreengrassSubscriptionDefinitionVersion, sd.Id, sd.Id, sd.Version, sd)
					sdRes.AddRelation(resource.GreengrassSubscriptionDefinition, v.Id, "")
					// TODO relationships to subscriptions
					r.AddRelation(resource.GreengrassSubscriptionDefinitionVersion, sd.Id, sd.Version)
					rg.AddResource(sdRes)
				}
				if subscriptionDefVersions.NextToken == nil {
					break
				}
				sdNextToken = subscriptionDefVersions.NextToken
			}
			rg.AddResource(r)
		}
		if subscriptiondefs.NextToken == nil {
			break
		}
		nextToken = subscriptiondefs.NextToken
	}
	return rg, nil
}
