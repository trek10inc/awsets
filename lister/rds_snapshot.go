package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSRdsDbSnapshot struct {
}

func init() {
	i := AWSRdsDbSnapshot{}
	listers = append(listers, i)
}

func (l AWSRdsDbSnapshot) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.RDSDbSnapshot}
}

func (l AWSRdsDbSnapshot) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := rds.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBSnapshots(ctx.Context, &rds.DescribeDBSnapshotsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.DBSnapshots {
			r := resource.New(ctx, resource.RDSDbSnapshot, v.DBSnapshotIdentifier, v.DBSnapshotIdentifier, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			r.AddRelation(resource.RdsDbInstance, v.DBInstanceIdentifier, "")
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")

			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
