package lister

import (
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
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

	res, err := svc.DescribeDBSnapshots(ctx.Context, &rds.DescribeDBSnapshotsInput{
		MaxRecords: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := rds.NewDescribeDBSnapshotsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.DBSnapshots {
			r := resource.New(ctx, resource.RDSDbSnapshot, v.DBSnapshotIdentifier, v.DBSnapshotIdentifier, v)
			if v.KmsKeyId != nil {
				kmsArn := arn.ParseP(v.KmsKeyId)
				r.AddRelation(resource.KmsKey, kmsArn.ResourceId, kmsArn.ResourceVersion)
			}
			r.AddRelation(resource.RdsDbInstance, v.DBInstanceIdentifier, "")
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
