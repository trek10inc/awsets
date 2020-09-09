package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/greengrass"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSGreengrassDeviceDefinition struct {
}

func init() {
	i := AWSGreengrassDeviceDefinition{}
	listers = append(listers, i)
}

func (l AWSGreengrassDeviceDefinition) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GreengrassDeviceDefinition,
		resource.GreengrassDeviceDefinitionVersion,
	}
}

func (l AWSGreengrassDeviceDefinition) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := greengrass.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		devicedefs, err := svc.ListDeviceDefinitionsRequest(&greengrass.ListDeviceDefinitionsInput{
			MaxResults: aws.String("100"),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			// greengrass errors are not of type awserr.Error
			if strings.Contains(err.Error(), "TooManyRequestsException") {
				// If greengrass is not supported in a region, returns "TooManyRequests exception"
				return rg, nil
			}
			return rg, fmt.Errorf("failed to list greengrass device definitions: %w", err)
		}
		for _, v := range devicedefs.Definitions {
			r := resource.New(ctx, resource.GreengrassDeviceDefinition, v.Id, v.Name, v)
			var ddNextToken *string
			for {
				deviceDefVersions, err := svc.ListDeviceDefinitionVersionsRequest(&greengrass.ListDeviceDefinitionVersionsInput{
					DeviceDefinitionId: v.Id,
					MaxResults:         aws.String("100"),
					NextToken:          ddNextToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list greengrass device definition versions for %s: %w", *v.Id, err)
				}
				for _, ddId := range deviceDefVersions.Versions {
					dd, err := svc.GetDeviceDefinitionVersionRequest(&greengrass.GetDeviceDefinitionVersionInput{
						DeviceDefinitionId:        ddId.Id,
						DeviceDefinitionVersionId: ddId.Version,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to list greengrass device definition version for %s, %s: %w", *ddId.Id, *ddId.Version, err)
					}
					ddRes := resource.NewVersion(ctx, resource.GreengrassDeviceDefinitionVersion, dd.Id, dd.Id, dd.Version, dd)
					ddRes.AddRelation(resource.GreengrassDeviceDefinition, v.Id, "")
					// TODO relationships to things
					r.AddRelation(resource.GreengrassDeviceDefinitionVersion, dd.Id, dd.Version)
					rg.AddResource(ddRes)
				}
				if deviceDefVersions.NextToken == nil {
					break
				}
				ddNextToken = deviceDefVersions.NextToken
			}
			rg.AddResource(r)
		}
		if devicedefs.NextToken == nil {
			break
		}
		nextToken = devicedefs.NextToken
	}
	return rg, nil
}
