package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/greengrass"
	"github.com/trek10inc/awsets/option"
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

func (l AWSGreengrassGroup) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := greengrass.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListGroups(cfg.Context, &greengrass.ListGroupsInput{
			MaxResults: aws.String("100"),
			NextToken:  nt,
		})
		if err != nil {
			// greengrass errors are not of type awserr.Error
			if strings.Contains(err.Error(), "TooManyRequestsException") {
				// If greengrass is not supported in a region, returns "TooManyRequests exception"
				return nil, nil
			}
			return nil, fmt.Errorf("failed to list greengrass groups: %w", err)
		}
		for _, v := range res.Groups {
			r := resource.New(cfg, resource.GreengrassGroup, v.Id, v.Name, v)

			// Versions
			err = Paginator(func(nt2 *string) (*string, error) {
				versions, err := svc.ListGroupVersions(cfg.Context, &greengrass.ListGroupVersionsInput{
					GroupId:    v.Id,
					MaxResults: aws.String("100"),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list greengrass group versions for %s: %w", *v.Id, err)
				}
				for _, gvId := range versions.Versions {
					gv, err := svc.GetGroupVersion(cfg.Context, &greengrass.GetGroupVersionInput{
						GroupId:        gvId.Id,
						GroupVersionId: gvId.Version,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to list greengrass group version for %s, %s: %w", *gvId.Id, *gvId.Version, err)
					}
					gvRes := resource.NewVersion(cfg, resource.GreengrassGroupVersion, gv.Id, gv.Id, gv.Version, gv)
					if def := gv.Definition; def != nil {
						// TODO relationships
					}
					gvRes.AddRelation(resource.GreengrassGroup, v.Id, "")
					r.AddRelation(resource.GreengrassGroupVersion, gv.Id, gv.Version)
					rg.AddResource(gvRes)
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
