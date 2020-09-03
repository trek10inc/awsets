package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSElasticacheParameterGroup struct {
}

func init() {
	i := AWSElasticacheParameterGroup{}
	listers = append(listers, i)
}

func (l AWSElasticacheParameterGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ElasticacheParameterGroup}
}

func (l AWSElasticacheParameterGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := elasticache.New(ctx.AWSCfg)

	req := svc.DescribeCacheParameterGroupsRequest(&elasticache.DescribeCacheParameterGroupsInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := elasticache.NewDescribeCacheParameterGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, group := range page.CacheParameterGroups {

			r := resource.New(ctx, resource.ElasticacheParameterGroup, group.CacheParameterGroupName, group.CacheParameterGroupName, group)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
