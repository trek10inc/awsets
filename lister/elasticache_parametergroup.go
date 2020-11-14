package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := elasticache.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeCacheParameterGroups(ctx.Context, &elasticache.DescribeCacheParameterGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, group := range res.CacheParameterGroups {

			r := resource.New(ctx, resource.ElasticacheParameterGroup, group.CacheParameterGroupName, group.CacheParameterGroupName, group)
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
