package lister

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSDMSReplicationSubnetGroup struct {
}

func init() {
	i := AWSDMSReplicationSubnetGroup{}
	listers = append(listers, i)
}

func (l AWSDMSReplicationSubnetGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DMSReplicationSubnetGroup}
}

func (l AWSDMSReplicationSubnetGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := databasemigrationservice.New(ctx.AWSCfg)

	req := svc.DescribeReplicationSubnetGroupsRequest(&databasemigrationservice.DescribeReplicationSubnetGroupsInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := databasemigrationservice.NewDescribeReplicationSubnetGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.ReplicationSubnetGroups {
			r := resource.New(ctx, resource.DMSReplicationSubnetGroup, v.ReplicationSubnetGroupIdentifier, v.ReplicationSubnetGroupIdentifier, v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			for _, sn := range v.Subnets {
				r.AddRelation(resource.Ec2Subnet, sn.SubnetIdentifier, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()

	if err != nil {
		if strings.Contains(err.Error(), "exceeded maximum number of attempts") {
			// If DMS is not supported in a region, it triggers this error
			err = nil
		}
	}
	return rg, err
}
