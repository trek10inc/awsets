package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSNeptuneDbClusterParameterGroup struct {
}

func init() {
	i := AWSNeptuneDbClusterParameterGroup{}
	listers = append(listers, i)
}

func (l AWSNeptuneDbClusterParameterGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.NeptuneDbClusterParameterGroup}
}

func (l AWSNeptuneDbClusterParameterGroup) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := neptune.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBClusterParameterGroups(cfg.Context, &neptune.DescribeDBClusterParameterGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list neptune cluster parameter groups: %w", err)
		}
		for _, v := range res.DBClusterParameterGroups {
			groupArn := arn.ParseP(v.DBClusterParameterGroupArn)
			r := resource.New(cfg, resource.NeptuneDbClusterParameterGroup, groupArn.ResourceId, "", v)
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
