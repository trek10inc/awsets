package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBProxies(ctx.Context, &rds.DescribeDBProxiesInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "InvalidAction") {
				// If DB Proxy is not supported in a region, returns an InvalidAction error
				return nil, nil
			}
			return nil, err
		}
		for _, v := range res.DBProxies {
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
			err = Paginator(func(nt2 *string) (*string, error) {
				targetGroups, err := svc.DescribeDBProxyTargetGroups(ctx.Context, &rds.DescribeDBProxyTargetGroupsInput{
					DBProxyName: v.DBProxyName,
					MaxRecords:  aws.Int32(100),
					Marker:      nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get target groups for %s: %w", *v.DBProxyName, err)
				}
				for _, tg := range targetGroups.TargetGroups {
					tgArn := arn.ParseP(tg.TargetGroupArn)
					tgR := resource.New(ctx, resource.RdsDbProxyTargetGroup, tgArn.ResourceId, tg.TargetGroupName, tg)
					tgR.AddRelation(resource.RdsDbProxy, proxyArn.ResourceId, "")
					rg.AddResource(tgR)
				}
				return targetGroups.Marker, nil
			})
			if err != nil {
				return nil, err
			}

			// DB Proxy Targets
			targets := make([]*types.DBProxyTarget, 0)
			err = Paginator(func(nt2 *string) (*string, error) {
				proxyTargets, err := svc.DescribeDBProxyTargets(ctx.Context, &rds.DescribeDBProxyTargetsInput{
					DBProxyName: v.DBProxyName,
					MaxRecords:  aws.Int32(100),
					Marker:      nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get proxy targets for %s: %w", *v.DBProxyName, err)
				}
				targets = append(targets, proxyTargets.Targets...)
				return proxyTargets.Marker, nil
			})
			if err != nil {
				return nil, err
			}
			if len(targets) > 0 {
				r.AddAttribute("ProxyTargets", targets)
			}
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
