package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSRdsDbClusterParameterGroup struct {
}

func init() {
	i := AWSRdsDbClusterParameterGroup{}
	listers = append(listers, i)
}

func (l AWSRdsDbClusterParameterGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.RdsDbClusterParameterGroup}
}

func (l AWSRdsDbClusterParameterGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := rds.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBClusterParameterGroups(ctx.Context, &rds.DescribeDBClusterParameterGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list cluster parameter groups: %w", err)
		}
		for _, pGroup := range res.DBClusterParameterGroups {
			groupArn := arn.ParseP(pGroup.DBClusterParameterGroupArn)
			r := resource.New(ctx, resource.RdsDbParameterGroup, groupArn.ResourceId, "", pGroup)
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
