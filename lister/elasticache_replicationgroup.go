package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSElasticacheReplicationGroup struct {
}

func init() {
	i := AWSElasticacheReplicationGroup{}
	listers = append(listers, i)
}

func (l AWSElasticacheReplicationGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ElasticacheReplicationGroup}
}

func (l AWSElasticacheReplicationGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := elasticache.New(ctx.AWSCfg)

	req := svc.DescribeReplicationGroupsRequest(&elasticache.DescribeReplicationGroupsInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := elasticache.NewDescribeReplicationGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, group := range page.ReplicationGroups {
			r := resource.New(ctx, resource.ElasticacheReplicationGroup, group.ReplicationGroupId, group.ReplicationGroupId, group)

			if group.KmsKeyId != nil && *group.KmsKeyId != "" {
				r.AddRelation(resource.KmsKey, group.KmsKeyId, "")
			}

			for _, mc := range group.MemberClusters {
				r.AddRelation(resource.ElasticacheCluster, mc, "")
			}

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
