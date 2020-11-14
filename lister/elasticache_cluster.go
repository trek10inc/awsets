package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSElasticacheCluster struct {
}

func init() {
	i := AWSElasticacheCluster{}
	listers = append(listers, i)
}

func (l AWSElasticacheCluster) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ElasticacheCluster}
}

func (l AWSElasticacheCluster) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := elasticache.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeCacheClusters(ctx.Context, &elasticache.DescribeCacheClustersInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, cluster := range res.CacheClusters {
			r := resource.New(ctx, resource.ElasticacheCluster, cluster.CacheClusterId, cluster.CacheClusterId, cluster)

			if cluster.CacheParameterGroup != nil {
				r.AddRelation(resource.ElasticacheParameterGroup, cluster.CacheParameterGroup.CacheParameterGroupName, "")
			}
			if cluster.CacheSubnetGroupName != nil {
				r.AddRelation(resource.ElasticacheSubnetGroup, cluster.CacheSubnetGroupName, "")
			}
			for _, sg := range cluster.SecurityGroups {
				//r.AddRelation(resourcetype.ElasticacheSecurityGroup, sg.SecurityGroupId, "")
				r.AddRelation(resource.Ec2SecurityGroup, sg.SecurityGroupId, "")
			}
			if cluster.ReplicationGroupId != nil {
				//sourceArn := arn.ParseP(cluster.ReplicationSourceIdentifier)
				r.AddRelation(resource.ElasticacheReplicationGroup, cluster.ReplicationGroupId, "")
			}
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
