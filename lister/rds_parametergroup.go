package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBParameterGroups(ctx.Context, &rds.DescribeDBParameterGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, pGroup := range res.DBParameterGroups {
			groupArn := arn.ParseP(pGroup.DBParameterGroupArn)
			r := resource.New(ctx, resource.RdsDbParameterGroup, groupArn.ResourceId, "", pGroup)
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
