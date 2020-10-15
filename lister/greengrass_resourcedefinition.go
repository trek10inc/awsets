package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/greengrass"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
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

func (l AWSGreengrassResourceDefinition) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := greengrass.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListResourceDefinitions(cfg.Context, &greengrass.ListResourceDefinitionsInput{
			MaxResults: aws.String("100"),
			NextToken:  nt,
		})
		if err != nil {
			// greengrass errors are not of type awserr.Error
			if strings.Contains(err.Error(), "TooManyRequestsException") {
				// If greengrass is not supported in a region, returns "TooManyRequests exception"
				return nil, nil
			}
			return nil, fmt.Errorf("failed to list greengrass resource definitions: %w", err)
		}
		for _, v := range res.Definitions {
			r := resource.New(cfg, resource.GreengrassGroup, v.Id, v.Name, v)

			// Versions
			err = Paginator(func(nt2 *string) (*string, error) {
				versions, err := svc.ListResourceDefinitionVersions(cfg.Context, &greengrass.ListResourceDefinitionVersionsInput{
					ResourceDefinitionId: v.Id,
					MaxResults:           aws.String("100"),
					NextToken:            nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list greengrass resource definition versions for %s: %w", *v.Id, err)
				}
				for _, rdId := range versions.Versions {
					rd, err := svc.GetResourceDefinitionVersion(cfg.Context, &greengrass.GetResourceDefinitionVersionInput{
						ResourceDefinitionId:        rdId.Id,
						ResourceDefinitionVersionId: rdId.Version,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to list greengrass resource definition version for %s, %s: %w", *rdId.Id, *rdId.Version, err)
					}
					rdRes := resource.NewVersion(cfg, resource.GreengrassResourceDefinitionVersion, rd.Id, rd.Id, rd.Version, rd)
					rdRes.AddRelation(resource.GreengrassResourceDefinition, v.Id, "")
					// TODO relationships to resources
					r.AddRelation(resource.GreengrassResourceDefinitionVersion, rd.Id, rd.Version)
					rg.AddResource(rdRes)
				}
				return res.NextToken, nil
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
