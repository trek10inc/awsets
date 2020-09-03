package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSDocDBInstance struct {
}

func init() {
	i := AWSDocDBInstance{}
	listers = append(listers, i)
}

func (l AWSDocDBInstance) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DocDBInstance}
}

func (l AWSDocDBInstance) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := docdb.New(ctx.AWSCfg)

	req := svc.DescribeDBInstancesRequest(&docdb.DescribeDBInstancesInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := docdb.NewDescribeDBInstancesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, instance := range page.DBInstances {
			r := resource.New(ctx, resource.DocDBInstance, instance.DBInstanceIdentifier, instance.DBInstanceIdentifier, instance)
			r.AddRelation(resource.DocDBCluster, instance.DBClusterIdentifier, "")
			r.AddRelation(resource.DocDBSubnetGroup, instance.DBSubnetGroup, "")
			for _, sg := range instance.VpcSecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, sg.VpcSecurityGroupId, "")
			}
			if instance.KmsKeyId != nil {
				kmsArn := arn.ParseP(instance.KmsKeyId)
				r.AddRelation(resource.KmsKey, kmsArn.ResourceId, "")
			}

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
