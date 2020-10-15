package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
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

func (l AWSKafkaCluster) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := kafka.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListClusters(cfg.Context, &kafka.ListClustersInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, cluster := range res.ClusterInfoList {
			clusterArn := arn.ParseP(cluster.ClusterArn)
			r := resource.New(cfg, resource.KafkaCluster, clusterArn.ResourceId, cluster.ClusterName, cluster)

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
