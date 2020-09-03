package lister

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	svc := elasticache.New(ctx.AWSCfg)

	req := svc.DescribeCacheSecurityGroupsRequest(&elasticache.DescribeCacheSecurityGroupsInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := elasticache.NewDescribeCacheSecurityGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, sg := range page.CacheSecurityGroups {
			r := resource.New(ctx, resource.ElasticacheSecurityGroup, sg.CacheSecurityGroupName, sg.CacheSecurityGroupName, sg)
			for _, ec2sg := range sg.EC2SecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, ec2sg.EC2SecurityGroupName, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	if aerr, ok := err.(awserr.Error); ok {
		if aerr.Code() == elasticache.ErrCodeInvalidParameterValueException &&
			strings.Contains(aerr.Message(), "Use of cache security groups is not permitted in this API version for your account.") {
			// NOTE: ElastiCache Security Groups are for use only when working with an ElastiCache cluster
			// outside of a VPC. If you are using a VPC, see the ElastiCache Subnet Group resource. Security groups are
			// only used in EC2-Classic set-ups, so ignore this error.
			err = nil
		}
	}
	return rg, err
}
