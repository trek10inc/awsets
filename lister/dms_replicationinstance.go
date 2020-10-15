package lister

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	"github.com/trek10inc/awsets/option"
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

func (l AWSDMSReplicationInstance) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := databasemigrationservice.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeReplicationInstances(cfg.Context, &databasemigrationservice.DescribeReplicationInstancesInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "exceeded maximum number of attempts") {
				// If DMS is not supported in a region, it triggers this error
				return nil, nil
			}
			return nil, err
		}
		for _, v := range res.ReplicationInstances {
			r := resource.New(cfg, resource.DMSReplicationInstance, v.ReplicationInstanceIdentifier, v.ReplicationInstanceIdentifier, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			r.AddRelation(resource.DMSReplicationSubnetGroup, v.ReplicationSubnetGroup, "")
			for _, sg := range v.VpcSecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, sg.VpcSecurityGroupId, "")
			}
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
