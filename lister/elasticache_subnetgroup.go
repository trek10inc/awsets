package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSElasticacheSubnetGroup struct {
}

func init() {
	i := AWSElasticacheSubnetGroup{}
	listers = append(listers, i)
}

func (l AWSElasticacheSubnetGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ElasticacheSubnetGroup}
}

func (l AWSElasticacheSubnetGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := elasticache.New(ctx.AWSCfg)

	req := svc.DescribeCacheSubnetGroupsRequest(&elasticache.DescribeCacheSubnetGroupsInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := elasticache.NewDescribeCacheSubnetGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, sg := range page.CacheSubnetGroups {
			r := resource.New(ctx, resource.ElasticacheSubnetGroup, sg.CacheSubnetGroupName, sg.CacheSubnetGroupName, sg)
			if sg.VpcId != nil && *sg.VpcId != "" {
				r.AddRelation(resource.Ec2Vpc, sg.VpcId, "")
			}
			for _, subnet := range sg.Subnets {
				r.AddRelation(resource.Ec2Subnet, subnet.SubnetIdentifier, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
