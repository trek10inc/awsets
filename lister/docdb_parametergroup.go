package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/docdb/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSDocDBParameterGroup struct {
}

func init() {
	i := AWSDocDBParameterGroup{}
	listers = append(listers, i)
}

func (l AWSDocDBParameterGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DocDBParameterGroup}
}

func (l AWSDocDBParameterGroup) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := docdb.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBClusterParameterGroups(cfg.Context, &docdb.DescribeDBClusterParameterGroupsInput{
			Marker:     nt,
			MaxRecords: aws.Int32(100),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get parameter groups: %w", err)
		}
		for _, group := range res.DBClusterParameterGroups {
			r := resource.New(cfg, resource.DocDBParameterGroup, group.DBClusterParameterGroupName, group.DBClusterParameterGroupName, group)

			var paramMarker *string
			parameterList := make([]*types.Parameter, 0)
			for {
				params, err := svc.DescribeDBClusterParameters(cfg.Context, &docdb.DescribeDBClusterParametersInput{
					DBClusterParameterGroupName: group.DBClusterParameterGroupName,
					Marker:                      paramMarker,
					MaxRecords:                  aws.Int32(100),
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get parameters for %s: %w", *group.DBClusterParameterGroupName, err)
				}
				for _, param := range params.Parameters {
					parameterList = append(parameterList, param)
				}
				if params.Marker == nil {
					break
				}
				paramMarker = params.Marker
			}
			r.AddAttribute("Parameters", parameterList)
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
