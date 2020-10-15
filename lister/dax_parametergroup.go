package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dax"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSDAXParameterGroup struct {
}

func init() {
	i := AWSDAXParameterGroup{}
	listers = append(listers, i)
}

func (l AWSDAXParameterGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DAXParameterGroup}
}

func (l AWSDAXParameterGroup) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := dax.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeParameterGroups(cfg.Context, &dax.DescribeParameterGroupsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "Access Denied to API Version: DAX_V3") {
				// Regions that don't support DAX return access denied
				return nil, nil
			}
			return nil, fmt.Errorf("failed to list dax parameter groups: %w", err)
		}
		for _, v := range res.ParameterGroups {
			r := resource.New(cfg, resource.DAXParameterGroup, v.ParameterGroupName, v.ParameterGroupName, v)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
