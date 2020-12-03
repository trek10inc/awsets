package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := ec2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeSnapshots(ctx.Context, &ec2.DescribeSnapshotsInput{
			Filters: []types.Filter{{
				Name:   aws.String("owner-id"),
				Values: []string{ctx.AccountId},
			}},
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.Snapshots {
			r := resource.New(ctx, resource.Ec2Snapshot, v.SnapshotId, v.SnapshotId, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			r.AddRelation(resource.Ec2Volume, v.VolumeId, "")
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
