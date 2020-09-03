package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/trek10inc/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

type AWSElbv2Loadbalancer struct {
}

func init() {
	i := AWSElbv2Loadbalancer{}
	listers = append(listers, i)
}

func (l AWSElbv2Loadbalancer) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ElbV2LoadBalancer, resource.ElbV2TargetGroup, resource.ElbV2Listener}
}

func (l AWSElbv2Loadbalancer) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := elasticloadbalancingv2.New(ctx.AWSCfg)

	req := svc.DescribeLoadBalancersRequest(&elasticloadbalancingv2.DescribeLoadBalancersInput{})

	rg := resource.NewGroup()
	paginator := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.LoadBalancers {
			lbArn := arn.ParseP(v.LoadBalancerArn)
			r := resource.New(ctx, resource.ElbV2LoadBalancer, lbArn.ResourceId, v.LoadBalancerName, v)

			if v.VpcId != nil && *v.VpcId != "" {
				r.AddRelation(resource.Ec2Vpc, *v.VpcId, "")
			}
			for _, sg := range v.SecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, sg, "")
			}
			rg.AddResource(r)

			tgReq := svc.DescribeTargetGroupsRequest(&elasticloadbalancingv2.DescribeTargetGroupsInput{
				LoadBalancerArn: v.LoadBalancerArn,
			})
			tgPaginator := elasticloadbalancingv2.NewDescribeTargetGroupsPaginator(tgReq)
			for tgPaginator.Next(ctx.Context) {
				tgPage := tgPaginator.CurrentPage()
				for _, tg := range tgPage.TargetGroups {
					tgArn := arn.ParseP(tg.TargetGroupArn)
					tgr := resource.New(ctx, resource.ElbV2TargetGroup, tgArn.ResourceId, tg.TargetGroupName, tg)
					tgr.AddRelation(resource.ElbV2LoadBalancer, v.LoadBalancerName, "")
					tgr.AddRelation(resource.Ec2Vpc, tg.VpcId, "")
					rg.AddResource(tgr)
				}
			}
			lReq := svc.DescribeListenersRequest(&elasticloadbalancingv2.DescribeListenersInput{
				LoadBalancerArn: v.LoadBalancerArn,
			})
			lPaginator := elasticloadbalancingv2.NewDescribeListenersPaginator(lReq)
			for lPaginator.Next(ctx.Context) {
				lPage := lPaginator.CurrentPage()
				for _, l := range lPage.Listeners {
					lArn := arn.ParseP(l.ListenerArn)
					lr := resource.New(ctx, resource.ElbV2Listener, lArn.ResourceId, lArn.ResourceId, l)
					lr.AddRelation(resource.ElbV2LoadBalancer, v.LoadBalancerName, "")
					for _, c := range l.Certificates {
						certArn := arn.ParseP(c.CertificateArn)
						lr.AddRelation(resource.AcmCertificate, certArn.ResourceId, certArn.ResourceVersion)
					}
					for _, a := range l.DefaultActions {
						// TODO: rel to user pool
						//if a.AuthenticateCognitoConfig != nil {}
						if a.TargetGroupArn != nil {
							tgArn := arn.ParseP(a.TargetGroupArn)
							lr.AddRelation(resource.ElbV2TargetGroup, tgArn.ResourceId, tgArn.ResourceVersion)
						}
					}

					var ruleMarker *string
					for {
						rules, err := svc.DescribeRulesRequest(&elasticloadbalancingv2.DescribeRulesInput{
							ListenerArn: l.ListenerArn,
							PageSize:    aws.Int64(100),
							Marker:      ruleMarker,
						}).Send(ctx.Context)
						if err != nil {
							return rg, fmt.Errorf("failed to list rules for listener %s: %w", *l.ListenerArn, err)
						}
						for _, rule := range rules.Rules {
							ruleArn := arn.ParseP(rule.RuleArn)
							rRes := resource.New(ctx, resource.ElbV2ListenerRule, ruleArn.ResourceId, ruleArn.ResourceId, rule)
							rRes.AddARNRelation(resource.ElbV2Listener, l.ListenerArn)
							rRes.AddARNRelation(resource.ElbV2LoadBalancer, l.LoadBalancerArn)
							rg.AddResource(rRes)
						}
						if rules.NextMarker == nil {
							break
						}
						ruleMarker = rules.NextMarker
					}

					rg.AddResource(lr)
				}
			}
		}
	}
	err := paginator.Err()
	return rg, err
}
