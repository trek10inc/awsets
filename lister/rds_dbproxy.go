package lister

import (
	"fmt"
	"strings"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/trek10inc/awsets/arn"
)

type AWSRdsDbProxy struct {
}

func init() {
	i := AWSRdsDbProxy{}
	listers = append(listers, i)
}

func (l AWSRdsDbProxy) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.RdsDbProxy,
		resource.RdsDbProxyTargetGroup,
	}
}

func (l AWSRdsDbProxy) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := rds.NewFromConfig(ctx.AWSCfg)

	res, err := svc.DescribeDBProxies(ctx.Context, &rds.DescribeDBProxiesInput{
		MaxRecords: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := rds.NewDescribeDBProxiesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.DBProxies {
			proxyArn := arn.ParseP(v.DBProxyArn)
			r := resource.New(ctx, resource.RdsDbProxy, proxyArn.ResourceId, v.DBProxyName, v)
			r.AddARNRelation(resource.IamRole, v.RoleArn)
			for _, sg := range v.VpcSecurityGroupIds {
				r.AddRelation(resource.Ec2SecurityGroup, sg, "")
			}
			for _, sn := range v.VpcSubnetIds {
				r.AddRelation(resource.Ec2Subnet, sn, "")
			}

			// DB Proxy Target Groups
			tgPaginator := rds.NewDescribeDBProxyTargetGroupsPaginator(svc.DescribeDBProxyTargetGroups(ctx.Context, &rds.DescribeDBProxyTargetGroupsInput{
				DBProxyName: v.DBProxyName,
				MaxRecords:  aws.Int32(100),
			}))
			for tgPaginator.Next(ctx.Context) {
				tgPage := tgPaginator.CurrentPage()
				for _, tg := range tgPage.TargetGroups {
					tgArn := arn.ParseP(tg.TargetGroupArn)
					tgR := resource.New(ctx, resource.RdsDbProxyTargetGroup, tgArn.ResourceId, tg.TargetGroupName, tg)
					tgR.AddRelation(resource.RdsDbProxy, proxyArn.ResourceId, "")
					rg.AddResource(tgR)
				}
			}
			err := tgPaginator.Err()
			if err != nil {
				return nil, fmt.Errorf("failed to get target groups for %s: %w", *v.DBProxyName, err)
			}

			// DB Proxy Targets
			targets := make([]rds.DBProxyTarget, 0)
			tPaginator := rds.NewDescribeDBProxyTargetsPaginator(svc.DescribeDBProxyTargets(ctx.Context, &rds.DescribeDBProxyTargetsInput{
				DBProxyName: v.DBProxyName,
				MaxRecords:  aws.Int32(100),
			}))
			for tPaginator.Next(ctx.Context) {
				tPage := tPaginator.CurrentPage()
				targets = append(targets, tPage.Targets...)
				// TODO: map to the actual instances
			}
		}
	}
	err := paginator.Err()
	if err != nil {
		if strings.Contains(err.Error(), "InvalidAction") {
			// If DB Proxy is not supported in a region, returns an InvalidAction error
			err = nil
		}
	}
	return rg, err
}
