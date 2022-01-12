package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := rds.NewFromConfig(ctx.AWSCfg)

	ignoredEngines := map[string]struct{}{
		"docdb":   {},
		"neptune": {},
	}

	rg := resource.NewGroup()

	paginator := rds.NewDescribeDBClustersPaginator(svc, &rds.DescribeDBClustersInput{
		MaxRecords: aws.Int32(100),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx.Context)
		if err != nil {
			return nil, err
		}
		for _, cluster := range page.DBClusters {
			if _, ok := ignoredEngines[*cluster.Engine]; ok {
				continue
			}

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
			r.AddARNRelation(resource.RdsDbInstance, cluster.ReplicationSourceIdentifier)
			for _, replica := range cluster.ReadReplicaIdentifiers {
				//TODO: cluster vs instance based on type?
				r.AddARNRelation(resource.RdsDbCluster, replica)
			}
			for _, role := range cluster.AssociatedRoles {
				r.AddARNRelation(resource.IamRole, role.RoleArn)
			}
			r.AddARNRelation(resource.KmsKey, cluster.KmsKeyId)

			rg.AddResource(r)
		}
	}
	return rg, nil
}
