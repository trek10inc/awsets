package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := elasticloadbalancingv2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeLoadBalancers(ctx.Context, &elasticloadbalancingv2.DescribeLoadBalancersInput{
			Marker: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.LoadBalancers {
			lbArn := arn.ParseP(v.LoadBalancerArn)
			r := resource.New(ctx, resource.ElbV2LoadBalancer, lbArn.ResourceId, v.LoadBalancerName, v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")

			for _, sg := range v.SecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, sg, "")
			}
			rg.AddResource(r)

			// Target Groups
			err = Paginator(func(nt2 *string) (*string, error) {
				targetGroups, err := svc.DescribeTargetGroups(ctx.Context, &elasticloadbalancingv2.DescribeTargetGroupsInput{
					LoadBalancerArn: v.LoadBalancerArn,
					Marker:          nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to describe target groups for %s: %w", v.LoadBalancerName, err)
				}
				for _, tg := range targetGroups.TargetGroups {
					tgArn := arn.ParseP(tg.TargetGroupArn)
					tgr := resource.New(ctx, resource.ElbV2TargetGroup, tgArn.ResourceId, tg.TargetGroupName, tg)
					tgr.AddRelation(resource.ElbV2LoadBalancer, v.LoadBalancerName, "")
					tgr.AddRelation(resource.Ec2Vpc, tg.VpcId, "")
					rg.AddResource(tgr)
				}
				return targetGroups.NextMarker, nil
			})
			if err != nil {
				return nil, err
			}

			// Listeners
			err = Paginator(func(nt2 *string) (*string, error) {
				listeners, err := svc.DescribeListeners(ctx.Context, &elasticloadbalancingv2.DescribeListenersInput{
					LoadBalancerArn: v.LoadBalancerArn,
					Marker:          nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to describe listeners for %s: %w", v.LoadBalancerName, err)
				}
				for _, l := range listeners.Listeners {
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

					err = Paginator(func(nt3 *string) (*string, error) {
						rules, err := svc.DescribeRules(ctx.Context, &elasticloadbalancingv2.DescribeRulesInput{
							ListenerArn: l.ListenerArn,
							PageSize:    aws.Int32(100),
							Marker:      nt3,
						})
						if err != nil {
							return nil, fmt.Errorf("failed to list rules for listener %s: %w", *l.ListenerArn, err)
						}
						for _, rule := range rules.Rules {
							ruleArn := arn.ParseP(rule.RuleArn)
							rRes := resource.New(ctx, resource.ElbV2ListenerRule, ruleArn.ResourceId, ruleArn.ResourceId, rule)
							rRes.AddARNRelation(resource.ElbV2Listener, l.ListenerArn)
							rRes.AddARNRelation(resource.ElbV2LoadBalancer, l.LoadBalancerArn)
							rg.AddResource(rRes)
						}
						return rules.NextMarker, nil
					})
					if err != nil {
						return nil, err
					}
					rg.AddResource(lr)
				}
				return listeners.NextMarker, nil
			})
			if err != nil {
				return nil, err
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
