package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := elasticache.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeCacheSubnetGroups(ctx.Context, &elasticache.DescribeCacheSubnetGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, sg := range res.CacheSubnetGroups {
			r := resource.New(ctx, resource.ElasticacheSubnetGroup, sg.CacheSubnetGroupName, sg.CacheSubnetGroupName, sg)
			if sg.VpcId != nil && *sg.VpcId != "" {
				r.AddRelation(resource.Ec2Vpc, sg.VpcId, "")
			}
			for _, subnet := range sg.Subnets {
				r.AddRelation(resource.Ec2Subnet, subnet.SubnetIdentifier, "")
			}
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
