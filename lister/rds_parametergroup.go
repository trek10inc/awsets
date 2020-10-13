package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/trek10inc/awsets/arn"
)

type AWSRdsDbParameterGroup struct {
}

func init() {
	i := AWSRdsDbParameterGroup{}
	listers = append(listers, i)
}

func (l AWSRdsDbParameterGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.RdsDbParameterGroup}
}

func (l AWSRdsDbParameterGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := rds.NewFromConfig(ctx.AWSCfg)

	res, err := svc.DescribeDBParameterGroups(ctx.Context, &rds.DescribeDBParameterGroupsInput{
		MaxRecords: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := rds.NewDescribeDBParameterGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, pGroup := range page.DBParameterGroups {
			groupArn := arn.ParseP(pGroup.DBParameterGroupArn)
			r := resource.New(ctx, resource.RdsDbParameterGroup, groupArn.ResourceId, "", pGroup)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
