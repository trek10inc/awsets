package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := elasticache.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeReplicationGroups(ctx.Context, &elasticache.DescribeReplicationGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, group := range res.ReplicationGroups {
			r := resource.New(ctx, resource.ElasticacheReplicationGroup, group.ReplicationGroupId, group.ReplicationGroupId, group)

			if group.KmsKeyId != nil && *group.KmsKeyId != "" {
				r.AddRelation(resource.KmsKey, group.KmsKeyId, "")
			}

			for _, mc := range group.MemberClusters {
				r.AddRelation(resource.ElasticacheCluster, mc, "")
			}

			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
