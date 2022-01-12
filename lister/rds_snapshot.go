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

	ignoredEngines := map[string]struct{}{
		"docdb":   {},
		"neptune": {},
	}

	rg := resource.NewGroup()

	paginator := rds.NewDescribeDBSnapshotsPaginator(svc, &rds.DescribeDBSnapshotsInput{
		MaxRecords: aws.Int32(100),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx.Context)
		if err != nil {
			return nil, err
		}
		for _, v := range page.DBSnapshots {
			if _, ok := ignoredEngines[*v.Engine]; ok {
				continue
			}
			r := resource.New(ctx, resource.RDSDbSnapshot, v.DBSnapshotIdentifier, v.DBSnapshotIdentifier, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			r.AddRelation(resource.RdsDbInstance, v.DBInstanceIdentifier, "")
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")

			rg.AddResource(r)
		}
	}
	return rg, nil
}
