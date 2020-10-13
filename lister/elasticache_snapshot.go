package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := elasticache.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeSnapshots(ctx.Context, &elasticache.DescribeSnapshotsInput{
			MaxRecords: aws.Int32(50),
			Marker:     nt,
		})

		if err != nil {
			return nil, err
		}
		for _, v := range res.Snapshots {
			r := resource.New(ctx, resource.ElasticacheSnapshot, v.SnapshotName, v.SnapshotName, v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			r.AddRelation(resource.KmsKey, v.KmsKeyId, "")
			r.AddRelation(resource.ElasticacheParameterGroup, v.CacheParameterGroupName, "")
			r.AddRelation(resource.ElasticacheSubnetGroup, v.CacheSubnetGroupName, "")
			r.AddRelation(resource.ElasticacheCluster, v.CacheClusterId, "")
			r.AddRelation(resource.ElasticacheReplicationGroup, v.ReplicationGroupId, "")

			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
