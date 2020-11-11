package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2SpotFleet struct {
}

func init() {
	i := AWSEc2SpotFleet{}
	listers = append(listers, i)
}

func (l AWSEc2SpotFleet) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.Ec2SpotFleet,
	}
}

func (l AWSEc2SpotFleet) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ec2.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeSpotFleetRequests(cfg.Context, &ec2.DescribeSpotFleetRequestsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.SpotFleetRequestConfigs {
			r := resource.New(cfg, resource.Ec2SpotFleet, v.SpotFleetRequestId, v.SpotFleetRequestId, v)
			if c := v.SpotFleetRequestConfig; c != nil {
				r.AddRelation(resource.IamRole, c.IamFleetRole, "")
				for _, ls := range c.LaunchSpecifications {
					r.AddRelation(resource.Ec2Subnet, ls.SubnetId, "")
					r.AddRelation(resource.Ec2Image, ls.ImageId, "")
					for _, sg := range ls.SecurityGroups {
						r.AddRelation(resource.Ec2SecurityGroup, sg, "")
					}
					r.AddRelation(resource.IamInstanceProfile, ls.IamInstanceProfile, "")
					for _, bd := range ls.BlockDeviceMappings {
						r.AddARNRelation(resource.KmsKey, bd.Ebs.KmsKeyId)
						r.AddRelation(resource.Ec2Snapshot, bd.Ebs.SnapshotId, "")
					}
				}
				for _, ltc := range c.LaunchTemplateConfigs {
					r.AddRelation(resource.Ec2LaunchTemplate, ltc.LaunchTemplateSpecification.LaunchTemplateId, ltc.LaunchTemplateSpecification.Version)
				}
				if c.LoadBalancersConfig != nil {
					if c.LoadBalancersConfig.ClassicLoadBalancersConfig != nil {
						for _, elb := range c.LoadBalancersConfig.ClassicLoadBalancersConfig.ClassicLoadBalancers {
							r.AddRelation(resource.ElbLoadBalancer, elb, "")
						}
					}
					if c.LoadBalancersConfig.TargetGroupsConfig != nil {
						for _, tg := range c.LoadBalancersConfig.TargetGroupsConfig.TargetGroups {
							r.AddARNRelation(resource.ElbV2TargetGroup, tg.Arn)
						}
					}
				}
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
