package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSDMSReplicationInstance struct {
}

func init() {
	i := AWSDMSReplicationInstance{}
	listers = append(listers, i)
}

func (l AWSDMSReplicationInstance) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DMSReplicationInstance}
}

func (l AWSDMSReplicationInstance) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := databasemigrationservice.New(ctx.AWSCfg)

	req := svc.DescribeReplicationInstancesRequest(&databasemigrationservice.DescribeReplicationInstancesInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := databasemigrationservice.NewDescribeReplicationInstancesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.ReplicationInstances {
			r := resource.New(ctx, resource.DMSReplicationInstance, v.ReplicationInstanceIdentifier, v.ReplicationInstanceIdentifier, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			r.AddRelation(resource.DMSReplicationSubnetGroup, v.ReplicationSubnetGroup, "")
			for _, sg := range v.VpcSecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, sg.VpcSecurityGroupId, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
