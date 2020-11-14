package lister

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSElasticacheSecurityGroup struct {
}

func init() {
	i := AWSElasticacheSecurityGroup{}
	listers = append(listers, i)
}

func (l AWSElasticacheSecurityGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ElasticacheSecurityGroup}
}

func (l AWSElasticacheSecurityGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := elasticache.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeCacheSecurityGroups(ctx.Context, &elasticache.DescribeCacheSecurityGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "Use of cache security groups is not permitted in this API version for your account.") {
				// NOTE: ElastiCache Security Groups are for use only when working with an ElastiCache cluster
				// outside of a VPC. If you are using a VPC, see the ElastiCache Subnet Group resource. Security groups are
				// only used in EC2-Classic set-ups, so ignore this error.
				return nil, nil
			}
			return nil, err
		}
		for _, sg := range res.CacheSecurityGroups {
			r := resource.New(ctx, resource.ElasticacheSecurityGroup, sg.CacheSecurityGroupName, sg.CacheSecurityGroupName, sg)
			for _, ec2sg := range sg.EC2SecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, ec2sg.EC2SecurityGroupName, "")
			}
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
