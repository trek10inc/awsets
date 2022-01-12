package lister

import (
	"strings"

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

	paginator := rds.NewDescribeDBClusterParameterGroupsPaginator(svc, &rds.DescribeDBClusterParameterGroupsInput{
		MaxRecords: aws.Int32(100),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx.Context)
		if err != nil {
			return nil, err
		}
		for _, pGroup := range page.DBClusterParameterGroups {
			if strings.Contains(*pGroup.DBParameterGroupFamily, "neptune") || strings.Contains(*pGroup.DBParameterGroupFamily, "docdb") {
				continue
			}
			groupArn := arn.ParseP(pGroup.DBClusterParameterGroupArn)
			r := resource.New(ctx, resource.RdsDbParameterGroup, groupArn.ResourceId, "", pGroup)
			rg.AddResource(r)
		}
	}
	return rg, nil
}
