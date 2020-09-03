package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
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

func (l AWSDocDBParameterGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := docdb.New(ctx.AWSCfg)

	rg := resource.NewGroup()
	var marker *string

	for {
		groups, err := svc.DescribeDBClusterParameterGroupsRequest(&docdb.DescribeDBClusterParameterGroupsInput{
			Marker:     marker,
			MaxRecords: aws.Int64(100),
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to get parameter groups: %w", err)
		}
		for _, group := range groups.DBClusterParameterGroups {
			r := resource.New(ctx, resource.DocDBParameterGroup, group.DBClusterParameterGroupName, group.DBClusterParameterGroupName, group)

			var paramMarker *string
			parameterList := make([]docdb.Parameter, 0)
			for {
				params, err := svc.DescribeDBClusterParametersRequest(&docdb.DescribeDBClusterParametersInput{
					DBClusterParameterGroupName: group.DBClusterParameterGroupName,
					Marker:                      paramMarker,
					MaxRecords:                  aws.Int64(100),
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to get parameters for %s: %w", *group.DBClusterParameterGroupName, err)
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
		if groups.Marker == nil {
			break
		}
		marker = groups.Marker
	}
	return rg, nil
}
