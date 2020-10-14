package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSRedshiftSnapshot struct {
}

func init() {
	i := AWSRedshiftSnapshot{}
	listers = append(listers, i)
}

func (l AWSRedshiftSnapshot) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.RedshiftSnapshot}
}

func (l AWSRedshiftSnapshot) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := redshift.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeClusterSnapshots(ctx.Context, &redshift.DescribeClusterSnapshotsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.Snapshots {
			r := resource.New(ctx, resource.RedshiftSnapshot, v.SnapshotIdentifier, v.SnapshotIdentifier, v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			r.AddRelation(resource.RedshiftCluster, v.ClusterIdentifier, "")

			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
