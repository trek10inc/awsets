package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSElasticacheSnapshot struct {
}

func init() {
	i := AWSElasticacheSnapshot{}
	listers = append(listers, i)
}

func (l AWSElasticacheSnapshot) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ElasticacheSnapshot}
}

func (l AWSElasticacheSnapshot) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := elasticache.New(ctx.AWSCfg)

	req := svc.DescribeSnapshotsRequest(&elasticache.DescribeSnapshotsInput{
		MaxRecords: aws.Int64(50),
	})

	rg := resource.NewGroup()
	paginator := elasticache.NewDescribeSnapshotsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Snapshots {
			r := resource.New(ctx, resource.ElasticacheSnapshot, v.SnapshotName, v.SnapshotName, v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			r.AddRelation(resource.KmsKey, v.KmsKeyId, "")
			r.AddRelation(resource.ElasticacheParameterGroup, v.CacheParameterGroupName, "")
			r.AddRelation(resource.ElasticacheSubnetGroup, v.CacheSubnetGroupName, "")
			r.AddRelation(resource.ElasticacheCluster, v.CacheClusterId, "")
			r.AddRelation(resource.ElasticacheReplicationGroup, v.ReplicationGroupId, "")

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
