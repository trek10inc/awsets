package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/greengrass"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
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

func (l AWSGreengrassSubscriptionDefinition) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := greengrass.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListSubscriptionDefinitions(cfg.Context, &greengrass.ListSubscriptionDefinitionsInput{
			MaxResults: aws.String("100"),
			NextToken:  nt,
		})
		if err != nil {
			// greengrass errors are not of type awserr.Error
			if strings.Contains(err.Error(), "TooManyRequestsException") {
				// If greengrass is not supported in a region, returns "TooManyRequests exception"
				return nil, nil
			}
			return nil, fmt.Errorf("failed to list greengrass subscription definitions: %w", err)
		}
		for _, v := range res.Definitions {
			r := resource.New(cfg, resource.GreengrassGroup, v.Id, v.Name, v)

			// Versions
			err = Paginator(func(nt2 *string) (*string, error) {
				versions, err := svc.ListSubscriptionDefinitionVersions(cfg.Context, &greengrass.ListSubscriptionDefinitionVersionsInput{
					SubscriptionDefinitionId: v.Id,
					MaxResults:               aws.String("100"),
					NextToken:                nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list greengrass subscription definition versions for %s: %w", *v.Id, err)
				}
				for _, sdId := range versions.Versions {
					sd, err := svc.GetSubscriptionDefinitionVersion(cfg.Context, &greengrass.GetSubscriptionDefinitionVersionInput{
						SubscriptionDefinitionId:        sdId.Id,
						SubscriptionDefinitionVersionId: sdId.Version,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to list greengrass subscription definition version for %s, %s: %w", *sdId.Id, *sdId.Version, err)
					}
					sdRes := resource.NewVersion(cfg, resource.GreengrassSubscriptionDefinitionVersion, sd.Id, sd.Id, sd.Version, sd)
					sdRes.AddRelation(resource.GreengrassSubscriptionDefinition, v.Id, "")
					// TODO relationships to subscriptions
					r.AddRelation(resource.GreengrassSubscriptionDefinitionVersion, sd.Id, sd.Version)
					rg.AddResource(sdRes)
				}
				return versions.NextToken, nil
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
