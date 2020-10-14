package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSNeptuneDbClusterSnapshot struct {
}

func init() {
	i := AWSNeptuneDbClusterSnapshot{}
	listers = append(listers, i)
}

func (l AWSNeptuneDbClusterSnapshot) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.NeptuneDbClusterSnapshot}
}

func (l AWSNeptuneDbClusterSnapshot) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := neptune.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBClusterSnapshots(ctx.Context, &neptune.DescribeDBClusterSnapshotsInput{
			Marker:     nt,
			MaxRecords: aws.Int32(100),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list neptune cluster snapshots: %w", err)
		}
		for _, v := range res.DBClusterSnapshots {
			r := resource.New(ctx, resource.NeptuneDbClusterSnapshot, v.DBClusterSnapshotIdentifier, v.DBClusterSnapshotIdentifier, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			r.AddRelation(resource.NeptuneDbCluster, v.DBClusterIdentifier, "")
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")

			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
