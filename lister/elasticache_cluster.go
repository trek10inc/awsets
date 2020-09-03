package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	svc := elasticache.New(ctx.AWSCfg)

	req := svc.DescribeCacheClustersRequest(&elasticache.DescribeCacheClustersInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := elasticache.NewDescribeCacheClustersPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, cluster := range page.CacheClusters {
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
	}
	err := paginator.Err()
	return rg, err
}
