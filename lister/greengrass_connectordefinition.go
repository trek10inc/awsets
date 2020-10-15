package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/greengrass"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSGreengrassConnectorDefinition struct {
}

func init() {
	i := AWSGreengrassConnectorDefinition{}
	listers = append(listers, i)
}

func (l AWSGreengrassConnectorDefinition) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GreengrassConnectorDefinition,
		resource.GreengrassConnectorDefinitionVersion,
	}
}

func (l AWSGreengrassConnectorDefinition) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := greengrass.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListConnectorDefinitions(cfg.Context, &greengrass.ListConnectorDefinitionsInput{
			MaxResults: aws.String("100"),
			NextToken:  nt,
		})
		if err != nil {
			// greengrass errors are not of type awserr.Error
			if strings.Contains(err.Error(), "TooManyRequestsException") {
				// If greengrass is not supported in a region, returns "TooManyRequests exception"
				return nil, nil
			}
			return nil, fmt.Errorf("failed to list greengrass connector definitions: %w", err)
		}
		for _, v := range res.Definitions {
			r := resource.New(cfg, resource.GreengrassConnectorDefinition, v.Id, v.Name, v)

			// Versions
			err = Paginator(func(nt2 *string) (*string, error) {
				versions, err := svc.ListConnectorDefinitionVersions(cfg.Context, &greengrass.ListConnectorDefinitionVersionsInput{
					ConnectorDefinitionId: v.Id,
					MaxResults:            aws.String("100"),
					NextToken:             nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list greengrass connector definition versions for %s: %w", *v.Id, err)
				}
				for _, cdId := range versions.Versions {
					cd, err := svc.GetConnectorDefinitionVersion(cfg.Context, &greengrass.GetConnectorDefinitionVersionInput{
						ConnectorDefinitionId:        cdId.Id,
						ConnectorDefinitionVersionId: cdId.Version,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to list greengrass connector definition version for %s, %s: %w", *cdId.Id, *cdId.Version, err)
					}
					cdRes := resource.NewVersion(cfg, resource.GreengrassConnectorDefinitionVersion, cd.Id, cd.Id, cd.Version, cd)
					cdRes.AddRelation(resource.GreengrassConnectorDefinition, v.Id, "")
					// TODO relationships to connectors
					r.AddRelation(resource.GreengrassConnectorDefinitionVersion, cd.Id, cd.Version)
					rg.AddResource(cdRes)
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
