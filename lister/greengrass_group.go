package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/greengrass"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSGreengrassGroup struct {
}

func init() {
	i := AWSGreengrassGroup{}
	listers = append(listers, i)
}

func (l AWSGreengrassGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GreengrassGroup,
		resource.GreengrassGroupVersion,
	}
}

func (l AWSGreengrassGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := greengrass.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		groups, err := svc.ListGroups(ctx.Context, &greengrass.ListGroupsInput{
			MaxResults: aws.String("100"),
			NextToken:  nextToken,
		})
		if err != nil {
			// greengrass errors are not of type awserr.Error
			if strings.Contains(err.Error(), "TooManyRequestsException") {
				// If greengrass is not supported in a region, returns "TooManyRequests exception"
				return rg, nil
			}
			return nil, fmt.Errorf("failed to list greengrass groups: %w", err)
		}
		for _, v := range groups.Groups {
			r := resource.New(ctx, resource.GreengrassGroup, v.Id, v.Name, v)
			var gvNextToken *string
			for {
				groupVersions, err := svc.ListGroupVersions(ctx.Context, &greengrass.ListGroupVersionsInput{
					GroupId:    v.Id,
					MaxResults: aws.String("100"),
					NextToken:  gvNextToken,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list greengrass group versions for %s: %w", *v.Id, err)
				}
				for _, gvId := range groupVersions.Versions {
					gv, err := svc.GetGroupVersion(ctx.Context, &greengrass.GetGroupVersionInput{
						GroupId:        gvId.Id,
						GroupVersionId: gvId.Version,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to list greengrass group version for %s, %s: %w", *gvId.Id, *gvId.Version, err)
					}
					gvRes := resource.NewVersion(ctx, resource.GreengrassGroupVersion, gv.Id, gv.Id, gv.Version, gv)
					if def := gv.Definition; def != nil {
						// TODO relationships
					}
					gvRes.AddRelation(resource.GreengrassGroup, v.Id, "")
					r.AddRelation(resource.GreengrassGroupVersion, gv.Id, gv.Version)
					rg.AddResource(gvRes)
				}
				if groupVersions.NextToken == nil {
					break
				}
				gvNextToken = groupVersions.NextToken
			}
			rg.AddResource(r)
		}
		if groups.NextToken == nil {
			break
		}
		nextToken = groups.NextToken
	}
	return rg, nil
}
