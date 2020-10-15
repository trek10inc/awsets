package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
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

func (l AWSDocDBInstance) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := docdb.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBInstances(cfg.Context, &docdb.DescribeDBInstancesInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, instance := range res.DBInstances {
			r := resource.New(cfg, resource.DocDBInstance, instance.DBInstanceIdentifier, instance.DBInstanceIdentifier, instance)
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
		return res.Marker, nil
	})
	return rg, err
}
