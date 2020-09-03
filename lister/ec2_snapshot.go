package lister

import (
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AWSEc2Snapshot struct {
}

func init() {
	i := AWSEc2Snapshot{}
	listers = append(listers, i)
}

func (l AWSEc2Snapshot) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2Snapshot}
}

func (l AWSEc2Snapshot) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeSnapshotsRequest(&ec2.DescribeSnapshotsInput{
		Filters: []ec2.Filter{{
			Name:   aws.String("owner-id"),
			Values: []string{ctx.AccountId},
		}},
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := ec2.NewDescribeSnapshotsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Snapshots {
			r := resource.New(ctx, resource.Ec2Snapshot, v.SnapshotId, v.SnapshotId, v)
			if v.KmsKeyId != nil {
				kmsArn := arn.ParseP(v.KmsKeyId)
				r.AddRelation(resource.KmsKey, kmsArn.ResourceId, kmsArn.ResourceVersion)
			}
			r.AddRelation(resource.Ec2Volume, v.VolumeId, "")
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
