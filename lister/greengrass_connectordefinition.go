package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/greengrass"
	"github.com/trek10inc/awsets/context"
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

func (l AWSGreengrassConnectorDefinition) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := greengrass.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		connectordefs, err := svc.ListConnectorDefinitionsRequest(&greengrass.ListConnectorDefinitionsInput{
			MaxResults: aws.String("100"),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			// greengrass errors are not of type awserr.Error
			if strings.Contains(err.Error(), "TooManyRequestsException") {
				// If greengrass is not supported in a region, returns "TooManyRequests exception"
				return rg, nil
			}
			return rg, fmt.Errorf("failed to list greengrass connector definitions: %w", err)
		}
		for _, v := range connectordefs.Definitions {
			r := resource.New(ctx, resource.GreengrassConnectorDefinition, v.Id, v.Name, v)
			var cdNextToken *string
			for {
				connectorDefVersions, err := svc.ListConnectorDefinitionVersionsRequest(&greengrass.ListConnectorDefinitionVersionsInput{
					ConnectorDefinitionId: v.Id,
					MaxResults:            aws.String("100"),
					NextToken:             cdNextToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list greengrass connector definition versions for %s: %w", *v.Id, err)
				}
				for _, cdId := range connectorDefVersions.Versions {
					cd, err := svc.GetConnectorDefinitionVersionRequest(&greengrass.GetConnectorDefinitionVersionInput{
						ConnectorDefinitionId:        cdId.Id,
						ConnectorDefinitionVersionId: cdId.Version,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to list greengrass connector definition version for %s, %s: %w", *cdId.Id, *cdId.Version, err)
					}
					cdRes := resource.NewVersion(ctx, resource.GreengrassConnectorDefinitionVersion, cd.Id, cd.Id, cd.Version, cd)
					cdRes.AddRelation(resource.GreengrassConnectorDefinition, v.Id, "")
					// TODO relationships to connectors
					r.AddRelation(resource.GreengrassConnectorDefinitionVersion, cd.Id, cd.Version)
					rg.AddResource(cdRes)
				}
				if connectorDefVersions.NextToken == nil {
					break
				}
				cdNextToken = connectorDefVersions.NextToken
			}
			rg.AddResource(r)
		}
		if connectordefs.NextToken == nil {
			break
		}
		nextToken = connectordefs.NextToken
	}
	return rg, nil
}
