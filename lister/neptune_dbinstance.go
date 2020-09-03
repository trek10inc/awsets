package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/neptune"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/arn"
)

type AWSNeptuneDbInstance struct {
}

func init() {
	i := AWSNeptuneDbInstance{}
	listers = append(listers, i)
}

func (l AWSNeptuneDbInstance) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.NeptuneDbInstance}
}

func (l AWSNeptuneDbInstance) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := neptune.New(ctx.AWSCfg)

	req := svc.DescribeDBInstancesRequest(&neptune.DescribeDBInstancesInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()

	paginator := neptune.NewDescribeDBInstancesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.DBInstances {
			dbArn := arn.ParseP(v.DBInstanceArn)
			r := resource.New(ctx, resource.NeptuneDbInstance, dbArn.ResourceId, "", v)
			for _, pgroup := range v.DBParameterGroups {
				r.AddRelation(resource.NeptuneDbParameterGroup, pgroup.DBParameterGroupName, "")
			}
			for _, sgroup := range v.DBSecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, sgroup.DBSecurityGroupName, "")
			}
			if v.DBSubnetGroup != nil {
				r.AddRelation(resource.Ec2Vpc, v.DBSubnetGroup.VpcId, "")
				if v.DBSubnetGroup.DBSubnetGroupArn != nil {
					subnetGroupArn := arn.ParseP(v.DBSubnetGroup.DBSubnetGroupArn)
					r.AddRelation(resource.NeptuneDbSubnetGroup, subnetGroupArn.ResourceId, subnetGroupArn.ResourceVersion)
				}
				for _, subnet := range v.DBSubnetGroup.Subnets {
					r.AddRelation(resource.Ec2Subnet, subnet.SubnetIdentifier, "")
				}
			}
			for _, vpcSg := range v.VpcSecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, vpcSg.VpcSecurityGroupId, "")
			}
			if v.ReadReplicaSourceDBInstanceIdentifier != nil {
				if arn.IsArnP(v.ReadReplicaSourceDBInstanceIdentifier) {
					ctx.Logger.Errorf("source db instance is actually an arn!: %s\n", *v.ReadReplicaSourceDBInstanceIdentifier)
				}
				r.AddRelation(resource.NeptuneDbInstance, *v.ReadReplicaSourceDBInstanceIdentifier, "")
			}
			for _, replicaCluster := range v.ReadReplicaDBClusterIdentifiers {
				clusterArn := arn.Parse(replicaCluster)
				r.AddRelation(resource.NeptuneDbCluster, clusterArn.ResourceId, clusterArn.ResourceVersion)
			}
			for _, replicaInstance := range v.ReadReplicaDBInstanceIdentifiers {
				r.AddRelation(resource.NeptuneDbInstance, replicaInstance, "")
			}
			if v.MonitoringRoleArn != nil {
				roleArn := arn.ParseP(v.MonitoringRoleArn)
				r.AddRelation(resource.IamRole, roleArn.ResourceId, roleArn.ResourceVersion)
			}
			if v.KmsKeyId != nil {
				kmsKeyArn := arn.ParseP(v.KmsKeyId)
				r.AddRelation(resource.KmsKey, kmsKeyArn.ResourceId, kmsKeyArn.ResourceVersion)
			}

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
