package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/neptune"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/arn"
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

func (l AWSNeptuneDbClusterParameterGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := neptune.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	var marker *string
	for {
		res, err := svc.DescribeDBClusterParameterGroups(ctx.Context, &neptune.DescribeDBClusterParameterGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     marker,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list neptune cluster parameter groups: %w", err)
		}
		for _, v := range res.DBClusterParameterGroups {
			groupArn := arn.ParseP(v.DBClusterParameterGroupArn)
			r := resource.New(ctx, resource.NeptuneDbClusterParameterGroup, groupArn.ResourceId, "", v)
			rg.AddResource(r)
		}
		if res.Marker == nil {
			break
		}
		marker = res.Marker
	}
	return rg, nil
}
