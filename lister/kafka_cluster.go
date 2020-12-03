package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSKafkaCluster struct {
}

func init() {
	i := AWSKafkaCluster{}
	listers = append(listers, i)
}

func (l AWSKafkaCluster) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.KafkaCluster}
}

func (l AWSKafkaCluster) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := kafka.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListClusters(ctx.Context, &kafka.ListClustersInput{
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, cluster := range res.ClusterInfoList {
			clusterArn := arn.ParseP(cluster.ClusterArn)
			r := resource.New(ctx, resource.KafkaCluster, clusterArn.ResourceId, cluster.ClusterName, cluster)

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
