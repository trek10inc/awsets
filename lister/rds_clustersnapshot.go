package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSRdsDbClusterSnapshot struct {
}

func init() {
	i := AWSRdsDbClusterSnapshot{}
	listers = append(listers, i)
}

func (l AWSRdsDbClusterSnapshot) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.RdsDbClusterSnapshot}
}

func (l AWSRdsDbClusterSnapshot) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := rds.NewFromConfig(ctx.AWSCfg)

	ignoredEngines := map[string]struct{}{
		"docdb":   {},
		"neptune": {},
	}

	rg := resource.NewGroup()

	paginator := rds.NewDescribeDBClusterSnapshotsPaginator(svc, &rds.DescribeDBClusterSnapshotsInput{
		MaxRecords: aws.Int32(100),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx.Context)
		if err != nil {
			return nil, fmt.Errorf("failed to list rds db cluster snapshots: %w", err)
		}
		for _, v := range page.DBClusterSnapshots {
			if _, ok := ignoredEngines[*v.Engine]; ok {
				continue
			}
			r := resource.New(ctx, resource.RdsDbClusterSnapshot, v.DBClusterSnapshotIdentifier, v.DBClusterSnapshotIdentifier, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			r.AddRelation(resource.RdsDbCluster, v.DBClusterIdentifier, "")
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")

			rg.AddResource(r)
		}
	}
	return rg, nil
}
