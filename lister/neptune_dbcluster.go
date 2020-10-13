package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/arn"
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

	var marker *string

	for {
		res, err := svc.DescribeDBClusters(ctx.Context, &neptune.DescribeDBClustersInput{
			Marker:     marker,
			MaxRecords: aws.Int32(100),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list neptune clusters: %w", err)
		}
		for _, cluster := range res.DBClusters {
			clusterArn := arn.ParseP(cluster.DBClusterArn)
			r := resource.New(ctx, resource.NeptuneDbCluster, clusterArn.ResourceId, "", cluster)

			r.AddRelation(resource.NeptuneDbClusterParameterGroup, cluster.DBClusterParameterGroup, "")
			r.AddRelation(resource.NeptuneDbSubnetGroup, cluster.DBSubnetGroup, "")
			for _, vpcSg := range cluster.VpcSecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, vpcSg.VpcSecurityGroupId, "")
			}
			if cluster.ReplicationSourceIdentifier != nil {
				sourceArn := arn.ParseP(cluster.ReplicationSourceIdentifier)
				r.AddRelation(resource.NeptuneDbInstance, sourceArn.ResourceId, sourceArn.ResourceVersion)
			}
			for _, replica := range cluster.ReadReplicaIdentifiers {
				//TODO: cluster vs instance based on type?
				replicaArn := arn.Parse(replica)
				r.AddRelation(resource.NeptuneDbCluster, replicaArn.ResourceId, replicaArn.ResourceVersion)
			}
			for _, role := range cluster.AssociatedRoles {
				roleArn := arn.ParseP(role.RoleArn)
				r.AddRelation(resource.IamRole, roleArn.ResourceId, roleArn.ResourceVersion)
			}
			r.AddARNRelation(resource.KmsKey, cluster.KmsKeyId)
			//if cluster.KmsKeyId != nil {
			//	kmsKeyArn := arn.ParseP(cluster.KmsKeyId)
			//	r.AddRelation(resource.KmsKey, kmsKeyArn.ResourceId, kmsKeyArn.ResourceVersion)
			//}

			rg.AddResource(r)
		}
		if res.Marker == nil {
			break
		}
		marker = res.Marker
	}
	return rg, nil
}
