package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSRdsDbInstance struct {
}

func init() {
	i := AWSRdsDbInstance{}
	listers = append(listers, i)
}

func (l AWSRdsDbInstance) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.RdsDbInstance}
}

func (l AWSRdsDbInstance) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := rds.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBInstances(ctx.Context, &rds.DescribeDBInstancesInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, dbInstance := range res.DBInstances {
			dbArn := arn.ParseP(dbInstance.DBInstanceArn)
			r := resource.New(ctx, resource.RdsDbInstance, dbArn.ResourceId, "", dbInstance)
			for _, pgroup := range dbInstance.DBParameterGroups {
				r.AddRelation(resource.RdsDbParameterGroup, pgroup.DBParameterGroupName, "")
			}
			for _, sgroup := range dbInstance.DBSecurityGroups {
				// TODO figure out distinction between Ec2SecurityGroups and DBSecurityGroups
				r.AddRelation(resource.Ec2SecurityGroup, sgroup.DBSecurityGroupName, "")
			}
			if dbInstance.DBSubnetGroup != nil {
				r.AddRelation(resource.Ec2Vpc, dbInstance.DBSubnetGroup.VpcId, "")
				if dbInstance.DBSubnetGroup.DBSubnetGroupArn != nil {
					subnetGroupArn := arn.ParseP(dbInstance.DBSubnetGroup.DBSubnetGroupArn)
					r.AddRelation(resource.RdsDbSubnetGroup, subnetGroupArn.ResourceId, subnetGroupArn.ResourceVersion)
				}
				for _, subnet := range dbInstance.DBSubnetGroup.Subnets {
					r.AddRelation(resource.Ec2Subnet, subnet.SubnetIdentifier, "")
				}
			}
			for _, vpcSg := range dbInstance.VpcSecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, vpcSg.VpcSecurityGroupId, "")
			}
			r.AddRelation(resource.RdsDbInstance, dbInstance.ReadReplicaSourceDBInstanceIdentifier, "")
			for _, replicaCluster := range dbInstance.ReadReplicaDBClusterIdentifiers {
				r.AddRelation(resource.RdsDbCluster, replicaCluster, "")
			}
			for _, replicaInstance := range dbInstance.ReadReplicaDBInstanceIdentifiers {
				r.AddRelation(resource.RdsDbInstance, replicaInstance, "")
			}
			for _, role := range dbInstance.AssociatedRoles {
				roleArn := arn.ParseP(role.RoleArn)
				r.AddRelation(resource.IamRole, roleArn.ResourceId, roleArn.ResourceVersion)
			}
			r.AddARNRelation(resource.IamRole, dbInstance.MonitoringRoleArn)
			r.AddARNRelation(resource.KmsKey, dbInstance.KmsKeyId)

			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
