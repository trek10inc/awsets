package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/greengrass"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSGreengrassLoggerDefinition struct {
}

func init() {
	i := AWSGreengrassLoggerDefinition{}
	listers = append(listers, i)
}

func (l AWSGreengrassLoggerDefinition) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.GreengrassLoggerDefinition}
}

func (l AWSGreengrassLoggerDefinition) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := greengrass.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		loggerdefs, err := svc.ListLoggerDefinitionsRequest(&greengrass.ListLoggerDefinitionsInput{
			MaxResults: aws.String("100"),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			// greengrass errors are not of type awserr.Error
			if strings.Contains(err.Error(), "TooManyRequestsException") {
				// If greengrass is not supported in a region, returns "TooManyRequests exception"
				return rg, nil
			}
			return rg, fmt.Errorf("failed to list greengrass logger definitions: %w", err)
		}
		for _, v := range loggerdefs.Definitions {
			r := resource.New(ctx, resource.GreengrassGroup, v.Id, v.Name, v)
			var ldNextToken *string
			for {
				loggerDefVersions, err := svc.ListLoggerDefinitionVersionsRequest(&greengrass.ListLoggerDefinitionVersionsInput{
					LoggerDefinitionId: v.Id,
					MaxResults:         aws.String("100"),
					NextToken:          ldNextToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list greengrass logger definition versions for %s: %w", *v.Id, err)
				}
				for _, ldId := range loggerDefVersions.Versions {
					ld, err := svc.GetLoggerDefinitionVersionRequest(&greengrass.GetLoggerDefinitionVersionInput{
						LoggerDefinitionId:        ldId.Id,
						LoggerDefinitionVersionId: ldId.Version,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to list greengrass logger definition version for %s, %s: %w", *ldId.Id, *ldId.Version, err)
					}
					ldRes := resource.NewVersion(ctx, resource.GreengrassLoggerDefinitionVersion, ld.Id, ld.Id, ld.Version, ld)
					ldRes.AddRelation(resource.GreengrassLoggerDefinition, v.Id, "")
					// TODO relationships to loggers
					r.AddRelation(resource.GreengrassLoggerDefinitionVersion, ld.Id, ld.Version)
					rg.AddResource(ldRes)
				}
				if loggerDefVersions.NextToken == nil {
					break
				}
				ldNextToken = loggerDefVersions.NextToken
			}
			rg.AddResource(r)
		}
		if loggerdefs.NextToken == nil {
			break
		}
		nextToken = loggerdefs.NextToken
	}
	return rg, nil
}
