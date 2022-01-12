package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/aws/aws-sdk-go-v2/service/neptune/types"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSNeptuneDbCluster struct {
}

func init() {
	i := AWSNeptuneDbCluster{}
	listers = append(listers, i)
}

func (l AWSNeptuneDbCluster) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.NeptuneDbCluster}
}

func (l AWSNeptuneDbCluster) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := neptune.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	paginator := neptune.NewDescribeDBClustersPaginator(svc, &neptune.DescribeDBClustersInput{
		MaxRecords: aws.Int32(100),
		Filters: []types.Filter{
			{
				Name:   aws.String("engine"),
				Values: []string{"neptune"},
			},
		},
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx.Context)
		if err != nil {
			return nil, fmt.Errorf("failed to list neptune clusters: %w", err)
		}
		for _, cluster := range page.DBClusters {
			clusterArn := arn.ParseP(cluster.DBClusterArn)
			r := resource.New(ctx, resource.NeptuneDbCluster, clusterArn.ResourceId, "", cluster)

			r.AddRelation(resource.NeptuneDbClusterParameterGroup, cluster.DBClusterParameterGroup, "")
			r.AddRelation(resource.NeptuneDbSubnetGroup, cluster.DBSubnetGroup, "")
			for _, vpcSg := range cluster.VpcSecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, vpcSg.VpcSecurityGroupId, "")
			}
			if cluster.ReplicationSourceIdentifier != nil {
				r.AddARNRelation(resource.NeptuneDbInstance, cluster.ReplicationSourceIdentifier)
			}
			for _, replica := range cluster.ReadReplicaIdentifiers {
				//TODO: cluster vs instance based on type?
				replicaArn := arn.Parse(replica)
				r.AddRelation(resource.NeptuneDbCluster, replicaArn.ResourceId, replicaArn.ResourceVersion)
			}
			for _, role := range cluster.AssociatedRoles {
				r.AddARNRelation(resource.IamRole, role.RoleArn)
			}
			r.AddARNRelation(resource.KmsKey, cluster.KmsKeyId)
			//if cluster.KmsKeyId != nil {
			//	kmsKeyArn := arn.ParseP(cluster.KmsKeyId)
			//	r.AddRelation(resource.KmsKey, kmsKeyArn.ResourceId, kmsKeyArn.ResourceVersion)
			//}

			rg.AddResource(r)
		}
	}
	return rg, nil
}
