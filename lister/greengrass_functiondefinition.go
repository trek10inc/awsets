package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/greengrass"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSGreengrassFunctionDefinition struct {
}

func init() {
	i := AWSGreengrassFunctionDefinition{}
	listers = append(listers, i)
}

func (l AWSGreengrassFunctionDefinition) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GreengrassFunctionDefinition,
		resource.GreengrassFunctionDefinitionVersion,
	}
}

func (l AWSGreengrassFunctionDefinition) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := greengrass.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		functiondefs, err := svc.ListFunctionDefinitionsRequest(&greengrass.ListFunctionDefinitionsInput{
			MaxResults: aws.String("100"),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			// greengrass errors are not of type awserr.Error
			if strings.Contains(err.Error(), "TooManyRequestsException") {
				// If greengrass is not supported in a region, returns "TooManyRequests exception"
				return rg, nil
			}
			return rg, fmt.Errorf("failed to list greengrass function definitions: %w", err)
		}
		for _, v := range functiondefs.Definitions {
			r := resource.New(ctx, resource.GreengrassFunctionDefinition, v.Id, v.Name, v)
			var fdNextToken *string
			for {
				functionDefVersions, err := svc.ListFunctionDefinitionVersionsRequest(&greengrass.ListFunctionDefinitionVersionsInput{
					FunctionDefinitionId: v.Id,
					MaxResults:           aws.String("100"),
					NextToken:            fdNextToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list greengrass function definition versions for %s: %w", *v.Id, err)
				}
				for _, fdId := range functionDefVersions.Versions {
					fd, err := svc.GetFunctionDefinitionVersionRequest(&greengrass.GetFunctionDefinitionVersionInput{
						FunctionDefinitionId:        fdId.Id,
						FunctionDefinitionVersionId: fdId.Version,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to list greengrass function definition version for %s, %s: %w", *fdId.Id, *fdId.Version, err)
					}
					fdRes := resource.NewVersion(ctx, resource.GreengrassFunctionDefinitionVersion, fd.Id, fd.Id, fd.Version, fd)
					fdRes.AddRelation(resource.GreengrassFunctionDefinition, v.Id, "")
					// TODO relationships to functions
					r.AddRelation(resource.GreengrassFunctionDefinitionVersion, fd.Id, fd.Version)
					rg.AddResource(fdRes)
				}
				if functionDefVersions.NextToken == nil {
					break
				}
				fdNextToken = functionDefVersions.NextToken
			}
			rg.AddResource(r)
		}
		if functiondefs.NextToken == nil {
			break
		}
		nextToken = functiondefs.NextToken
	}
	return rg, nil
}
