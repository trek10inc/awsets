package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/greengrass"
	"github.com/trek10inc/awsets/option"
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

func (l AWSGreengrassFunctionDefinition) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := greengrass.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListFunctionDefinitions(cfg.Context, &greengrass.ListFunctionDefinitionsInput{
			MaxResults: aws.String("100"),
			NextToken:  nt,
		})
		if err != nil {
			// greengrass errors are not of type awserr.Error
			if strings.Contains(err.Error(), "TooManyRequestsException") {
				// If greengrass is not supported in a region, returns "TooManyRequests exception"
				return nil, nil
			}
			return nil, fmt.Errorf("failed to list greengrass function definitions: %w", err)
		}
		for _, v := range res.Definitions {
			r := resource.New(cfg, resource.GreengrassFunctionDefinition, v.Id, v.Name, v)

			// Versions
			err = Paginator(func(nt2 *string) (*string, error) {
				versions, err := svc.ListFunctionDefinitionVersions(cfg.Context, &greengrass.ListFunctionDefinitionVersionsInput{
					FunctionDefinitionId: v.Id,
					MaxResults:           aws.String("100"),
					NextToken:            nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list greengrass function definition versions for %s: %w", *v.Id, err)
				}
				for _, fdId := range versions.Versions {
					fd, err := svc.GetFunctionDefinitionVersion(cfg.Context, &greengrass.GetFunctionDefinitionVersionInput{
						FunctionDefinitionId:        fdId.Id,
						FunctionDefinitionVersionId: fdId.Version,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to list greengrass function definition version for %s, %s: %w", *fdId.Id, *fdId.Version, err)
					}
					fdRes := resource.NewVersion(cfg, resource.GreengrassFunctionDefinitionVersion, fd.Id, fd.Id, fd.Version, fd)
					fdRes.AddRelation(resource.GreengrassFunctionDefinition, v.Id, "")
					// TODO relationships to functions
					r.AddRelation(resource.GreengrassFunctionDefinitionVersion, fd.Id, fd.Version)
					rg.AddResource(fdRes)
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
