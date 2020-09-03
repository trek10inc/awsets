package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
)

type AWSElbLoadbalancer struct {
}

func init() {
	i := AWSElbLoadbalancer{}
	listers = append(listers, i)
}

func (l AWSElbLoadbalancer) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ElbLoadBalancer}
}

func (l AWSElbLoadbalancer) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := elasticloadbalancing.New(ctx.AWSCfg)

	req := svc.DescribeLoadBalancersRequest(&elasticloadbalancing.DescribeLoadBalancersInput{})

	rg := resource.NewGroup()
	paginator := elasticloadbalancing.NewDescribeLoadBalancersPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.LoadBalancerDescriptions {
			r := resource.New(ctx, resource.ElbLoadBalancer, v.LoadBalancerName, v.LoadBalancerName, v)

			if v.VPCId != nil && *v.VPCId != "" {
				r.AddRelation(resource.Ec2Vpc, *v.VPCId, "")
			}
			for _, i := range v.Instances {
				r.AddRelation(resource.Ec2Instance, *i.InstanceId, "")
			}
			for _, s := range v.Subnets {
				r.AddRelation(resource.Ec2Subnet, s, "")
			}
			for _, sg := range v.SecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, sg, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
