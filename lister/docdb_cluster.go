package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSDocDBCluster struct {
}

func init() {
	i := AWSDocDBCluster{}
	listers = append(listers, i)
}

func (l AWSDocDBCluster) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DocDBCluster}
}

func (l AWSDocDBCluster) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := docdb.New(ctx.AWSCfg)

	req := svc.DescribeDBClustersRequest(&docdb.DescribeDBClustersInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := docdb.NewDescribeDBClustersPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, cluster := range page.DBClusters {

			r := resource.New(ctx, resource.DocDBCluster, cluster.DBClusterIdentifier, cluster.DBClusterIdentifier, cluster)
			r.AddRelation(resource.DocDBSubnetGroup, cluster.DBSubnetGroup, "")
			r.AddRelation(resource.DocDBParameterGroup, cluster.DBClusterParameterGroup, "")
			for _, sg := range cluster.VpcSecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, sg.VpcSecurityGroupId, "")
			}
			for _, role := range cluster.AssociatedRoles {
				roleArn := arn.ParseP(role.RoleArn)
				r.AddRelation(resource.IamRole, roleArn.ResourceId, "")
			}
			if cluster.KmsKeyId != nil {
				kmsArn := arn.ParseP(cluster.KmsKeyId)
				r.AddRelation(resource.KmsKey, kmsArn.ResourceId, "")
			}

			for _, member := range cluster.DBClusterMembers {
				r.AddRelation(resource.DocDBInstance, member.DBInstanceIdentifier, "")
			}

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
