package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
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

func (l AWSRdsDbClusterParameterGroup) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := rds.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBClusterParameterGroups(cfg.Context, &rds.DescribeDBClusterParameterGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list cluster parameter groups: %w", err)
		}
		for _, pGroup := range res.DBClusterParameterGroups {
			groupArn := arn.ParseP(pGroup.DBClusterParameterGroupArn)
			r := resource.New(cfg, resource.RdsDbParameterGroup, groupArn.ResourceId, "", pGroup)
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
