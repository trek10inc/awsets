package lister

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := neptune.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	paginator := neptune.NewDescribeDBParameterGroupsPaginator(svc, &neptune.DescribeDBParameterGroupsInput{
		MaxRecords: aws.Int32(100),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx.Context)
		if err != nil {
			return nil, err
		}
		for _, v := range page.DBParameterGroups {
			if !strings.Contains(*v.DBParameterGroupFamily, "neptune") {
				continue
			}
			groupArn := arn.ParseP(v.DBParameterGroupArn)
			r := resource.New(ctx, resource.NeptuneDbParameterGroup, groupArn.ResourceId, "", v)
			rg.AddResource(r)
		}
	}
	return rg, nil
}
