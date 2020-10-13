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
	svc := docdb.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBClusters(ctx.Context, &docdb.DescribeDBClustersInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, cluster := range res.DBClusters {
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
		return res.Marker, nil
	})
	return rg, err
}
