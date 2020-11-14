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

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBClusterSnapshots(ctx.Context, &rds.DescribeDBClusterSnapshotsInput{
			Marker:     nt,
			MaxRecords: aws.Int32(100),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list rds db cluster snapshots: %w", err)
		}

		for _, v := range res.DBClusterSnapshots {
			r := resource.New(ctx, resource.RdsDbClusterSnapshot, v.DBClusterSnapshotIdentifier, v.DBClusterSnapshotIdentifier, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			r.AddRelation(resource.RdsDbCluster, v.DBClusterIdentifier, "")
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")

			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
