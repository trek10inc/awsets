package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/arn"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
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

	rg := resource.NewGroup()
	var marker *string
	for {
		snapshots, err := svc.DescribeDBClusterSnapshots(ctx.Context, &rds.DescribeDBClusterSnapshotsInput{
			Marker:     marker,
			MaxRecords: aws.Int32(100),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list rds db cluster snapshots: %w", err)
		}

		for _, v := range snapshots.DBClusterSnapshots {
			r := resource.New(ctx, resource.RdsDbClusterSnapshot, v.DBClusterSnapshotIdentifier, v.DBClusterSnapshotIdentifier, v)
			if v.KmsKeyId != nil {
				kmsArn := arn.ParseP(v.KmsKeyId)
				r.AddRelation(resource.KmsKey, kmsArn.ResourceId, kmsArn.ResourceVersion)
			}
			r.AddRelation(resource.RdsDbCluster, v.DBClusterIdentifier, "")
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")

			rg.AddResource(r)
		}
		if snapshots.Marker == nil {
			break
		}
		marker = snapshots.Marker
	}

	return rg, nil
}
