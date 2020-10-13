package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/trek10inc/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/aws"
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

	res, err := svc.ListClusters(ctx.Context, &kafka.ListClustersInput{
		MaxResults: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := kafka.NewListClustersPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, cluster := range page.ClusterInfoList {
			clusterArn := arn.ParseP(cluster.ClusterArn)
			r := resource.New(ctx, resource.KafkaCluster, clusterArn.ResourceId, cluster.ClusterName, cluster)

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
