package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := elasticloadbalancing.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeLoadBalancers(ctx.Context, &elasticloadbalancing.DescribeLoadBalancersInput{
			Marker: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.LoadBalancerDescriptions {
			r := resource.New(ctx, resource.ElbLoadBalancer, v.LoadBalancerName, v.LoadBalancerName, v)

			if v.VPCId != nil && *v.VPCId != "" {
				r.AddRelation(resource.Ec2Vpc, v.VPCId, "")
			}
			for _, i := range v.Instances {
				r.AddRelation(resource.Ec2Instance, i.InstanceId, "")
			}
			for _, s := range v.Subnets {
				r.AddRelation(resource.Ec2Subnet, s, "")
			}
			for _, sg := range v.SecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, sg, "")
			}
			rg.AddResource(r)
		}
		return res.NextMarker, nil
	})
	return rg, err
}
