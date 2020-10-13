package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/greengrass"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSGreengrassCoreDefinition struct {
}

func init() {
	i := AWSGreengrassCoreDefinition{}
	listers = append(listers, i)
}

func (l AWSGreengrassCoreDefinition) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GreengrassCoreDefinition,
		resource.GreengrassCoreDefinitionVersion,
	}
}

func (l AWSGreengrassCoreDefinition) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := greengrass.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		coredefs, err := svc.ListCoreDefinitions(ctx.Context, &greengrass.ListCoreDefinitionsInput{
			MaxResults: aws.String("100"),
			NextToken:  nextToken,
		})
		if err != nil {
			// greengrass errors are not of type awserr.Error
			if strings.Contains(err.Error(), "TooManyRequestsException") {
				// If greengrass is not supported in a region, returns "TooManyRequests exception"
				return rg, nil
			}
			return nil, fmt.Errorf("failed to list greengrass core definitions: %w", err)
		}
		for _, v := range coredefs.Definitions {
			r := resource.New(ctx, resource.GreengrassCoreDefinition, v.Id, v.Name, v)
			var cdNextToken *string
			for {
				coreDefVersions, err := svc.ListCoreDefinitionVersions(ctx.Context, &greengrass.ListCoreDefinitionVersionsInput{
					CoreDefinitionId: v.Id,
					MaxResults:       aws.String("100"),
					NextToken:        cdNextToken,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list greengrass core definition versions for %s: %w", *v.Id, err)
				}
				for _, cdId := range coreDefVersions.Versions {
					cd, err := svc.GetCoreDefinitionVersion(ctx.Context, &greengrass.GetCoreDefinitionVersionInput{
						CoreDefinitionId:        cdId.Id,
						CoreDefinitionVersionId: cdId.Version,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to list greengrass core definition version for %s, %s: %w", *cdId.Id, *cdId.Version, err)
					}
					cdRes := resource.NewVersion(ctx, resource.GreengrassCoreDefinitionVersion, cd.Id, cd.Id, cd.Version, cd)
					cdRes.AddRelation(resource.GreengrassCoreDefinition, v.Id, "")
					// TODO relationships to things
					r.AddRelation(resource.GreengrassCoreDefinitionVersion, cd.Id, cd.Version)
					rg.AddResource(cdRes)
				}
				if coreDefVersions.NextToken == nil {
					break
				}
				cdNextToken = coreDefVersions.NextToken
			}
			rg.AddResource(r)
		}
		if coredefs.NextToken == nil {
			break
		}
		nextToken = coredefs.NextToken
	}
	return rg, nil
}
