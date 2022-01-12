package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/aws/aws-sdk-go-v2/service/neptune/types"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := neptune.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	paginator := neptune.NewDescribeDBInstancesPaginator(svc, &neptune.DescribeDBInstancesInput{
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
			return nil, err
		}
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
					r.AddARNRelation(resource.NeptuneDbSubnetGroup, v.DBSubnetGroup.DBSubnetGroupArn)
				}
				for _, subnet := range v.DBSubnetGroup.Subnets {
					r.AddRelation(resource.Ec2Subnet, subnet.SubnetIdentifier, "")
				}
			}
			for _, vpcSg := range v.VpcSecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, vpcSg.VpcSecurityGroupId, "")
			}
			r.AddARNRelation(resource.NeptuneDbInstance, v.ReadReplicaSourceDBInstanceIdentifier)
			for _, replicaCluster := range v.ReadReplicaDBClusterIdentifiers {
				r.AddRelation(resource.NeptuneDbCluster, replicaCluster, "")
			}
			for _, replicaInstance := range v.ReadReplicaDBInstanceIdentifiers {
				r.AddARNRelation(resource.NeptuneDbInstance, replicaInstance)
			}
			r.AddARNRelation(resource.IamRole, v.MonitoringRoleArn)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)

			rg.AddResource(r)
		}
	}
	return rg, nil
}
