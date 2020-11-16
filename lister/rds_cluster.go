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

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBClusters(ctx.Context, &rds.DescribeDBClustersInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, cluster := range res.DBClusters {
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
		return res.Marker, nil
	})
	return rg, err
}
