package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/trek10inc/awsets/arn"
)

type AWSRdsDbCluster struct {
}

func init() {
	i := AWSRdsDbCluster{}
	listers = append(listers, i)
}

func (l AWSRdsDbCluster) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.RdsDbCluster}
}

func (l AWSRdsDbCluster) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := rds.New(ctx.AWSCfg)

	req := svc.DescribeDBClustersRequest(&rds.DescribeDBClustersInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := rds.NewDescribeDBClustersPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, cluster := range page.DBClusters {
			clusterArn := arn.ParseP(cluster.DBClusterArn)
			r := resource.New(ctx, resource.RdsDbCluster, clusterArn.ResourceId, "", cluster)

			if cluster.DBClusterParameterGroup != nil {
				r.AddRelation(resource.RdsDbParameterGroup, cluster.DBClusterParameterGroup, "")
			}
			if cluster.DBSubnetGroup != nil {
				r.AddRelation(resource.RdsDbSubnetGroup, cluster.DBSubnetGroup, "")
			}
			for _, vpcSg := range cluster.VpcSecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, vpcSg.VpcSecurityGroupId, "")
			}
			if cluster.ReplicationSourceIdentifier != nil {
				sourceArn := arn.ParseP(cluster.ReplicationSourceIdentifier)
				r.AddRelation(resource.RdsDbInstance, sourceArn.ResourceId, sourceArn.ResourceVersion)
			}
			for _, replica := range cluster.ReadReplicaIdentifiers {
				//TODO: cluster vs instance based on type?
				replicaArn := arn.Parse(replica)
				r.AddRelation(resource.RdsDbCluster, replicaArn.ResourceId, replicaArn.ResourceVersion)
			}
			for _, role := range cluster.AssociatedRoles {
				roleArn := arn.ParseP(role.RoleArn)
				r.AddRelation(resource.IamRole, roleArn.ResourceId, roleArn.ResourceVersion)
			}
			if cluster.KmsKeyId != nil {
				kmsKeyArn := arn.ParseP(cluster.KmsKeyId)
				r.AddRelation(resource.KmsKey, kmsKeyArn.ResourceId, kmsKeyArn.ResourceVersion)
			}

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
