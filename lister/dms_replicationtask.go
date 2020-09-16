package lister

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSDMSReplicationTask struct {
}

func init() {
	i := AWSDMSReplicationTask{}
	listers = append(listers, i)
}

func (l AWSDMSReplicationTask) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DMSReplicationTask}
}

func (l AWSDMSReplicationTask) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := databasemigrationservice.New(ctx.AWSCfg)

	req := svc.DescribeReplicationTasksRequest(&databasemigrationservice.DescribeReplicationTasksInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := databasemigrationservice.NewDescribeReplicationTasksPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.ReplicationTasks {
			r := resource.New(ctx, resource.DMSReplicationTask, v.ReplicationTaskIdentifier, v.ReplicationTaskIdentifier, v)
			r.AddARNRelation(resource.DMSReplicationInstance, v.ReplicationInstanceArn)
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
