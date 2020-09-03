package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/greengrass"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSGreengrassResourceDefinition struct {
}

func init() {
	i := AWSGreengrassResourceDefinition{}
	listers = append(listers, i)
}

func (l AWSGreengrassResourceDefinition) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GreengrassResourceDefinition,
		resource.GreengrassResourceDefinitionVersion,
	}
}

func (l AWSGreengrassResourceDefinition) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := greengrass.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		resourcedefs, err := svc.ListResourceDefinitionsRequest(&greengrass.ListResourceDefinitionsInput{
			MaxResults: aws.String("100"),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list greengrass resource definitions: %w", err)
		}
		for _, v := range resourcedefs.Definitions {
			r := resource.New(ctx, resource.GreengrassGroup, v.Id, v.Name, v)
			var cdNextToken *string
			for {
				resourceDefVersions, err := svc.ListResourceDefinitionVersionsRequest(&greengrass.ListResourceDefinitionVersionsInput{
					ResourceDefinitionId: v.Id,
					MaxResults:           aws.String("100"),
					NextToken:            cdNextToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list greengrass resource definition versions for %s: %w", *v.Id, err)
				}
				for _, rdId := range resourceDefVersions.Versions {
					rd, err := svc.GetResourceDefinitionVersionRequest(&greengrass.GetResourceDefinitionVersionInput{
						ResourceDefinitionId:        rdId.Id,
						ResourceDefinitionVersionId: rdId.Version,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to list greengrass resource definition version for %s, %s: %w", *rdId.Id, *rdId.Version, err)
					}
					rdRes := resource.NewVersion(ctx, resource.GreengrassResourceDefinitionVersion, rd.Id, rd.Id, rd.Version, rd)
					rdRes.AddRelation(resource.GreengrassResourceDefinition, v.Id, "")
					// TODO relationships to resources
					r.AddRelation(resource.GreengrassResourceDefinitionVersion, rd.Id, rd.Version)
					rg.AddResource(rdRes)
				}
				if resourceDefVersions.NextToken == nil {
					break
				}
				cdNextToken = resourceDefVersions.NextToken
			}
			rg.AddResource(r)
		}
		if resourcedefs.NextToken == nil {
			break
		}
		nextToken = resourcedefs.NextToken
	}
	return rg, nil
}
