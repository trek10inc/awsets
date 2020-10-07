package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/redshift"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	svc := redshift.New(ctx.AWSCfg)

	req := svc.DescribeClusterSnapshotsRequest(&redshift.DescribeClusterSnapshotsInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := redshift.NewDescribeClusterSnapshotsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Snapshots {
			r := resource.New(ctx, resource.RedshiftSnapshot, v.SnapshotIdentifier, v.SnapshotIdentifier, v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			r.AddRelation(resource.RedshiftCluster, v.ClusterIdentifier, "")

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
