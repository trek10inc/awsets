package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/arn"
)

type AWSNeptuneDbParameterGroup struct {
}

func init() {
	i := AWSNeptuneDbParameterGroup{}
	listers = append(listers, i)
}

func (l AWSNeptuneDbParameterGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.NeptuneDbParameterGroup}
}

func (l AWSNeptuneDbParameterGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := neptune.New(ctx.AWSCfg)

	req := svc.DescribeDBParameterGroupsRequest(&neptune.DescribeDBParameterGroupsInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := neptune.NewDescribeDBParameterGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.DBParameterGroups {
			groupArn := arn.ParseP(v.DBParameterGroupArn)
			r := resource.New(ctx, resource.NeptuneDbParameterGroup, groupArn.ResourceId, "", v)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
