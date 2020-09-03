package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/arn"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
)

type AWSEfsFileSystems struct {
}

func init() {
	i := AWSEfsFileSystems{}
	listers = append(listers, i)
}

func (l AWSEfsFileSystems) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.EfsFileSystem, resource.EfsMountTarget}
}

func (l AWSEfsFileSystems) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := efs.New(ctx.AWSCfg)

	rg := resource.NewGroup()

	var fsMarker *string
	for {
		res, err := svc.DescribeFileSystemsRequest(&efs.DescribeFileSystemsInput{
			Marker:   fsMarker,
			MaxItems: aws.Int64(10),
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to describe efs file systems: %w", err)
		}

		for _, fs := range res.FileSystems {
			r := resource.New(ctx, resource.EfsFileSystem, fs.FileSystemId, fs.Name, fs)
			if fs.KmsKeyId != nil {
				kmsArn := arn.ParseP(fs.KmsKeyId)
				r.AddRelation(resource.KmsKey, kmsArn.ResourceId, "")
			}
			rg.AddResource(r)
			var mtMarker *string
			for {
				mtRes, err := svc.DescribeMountTargetsRequest(&efs.DescribeMountTargetsInput{
					FileSystemId: fs.FileSystemId,
					Marker:       mtMarker,
					MaxItems:     aws.Int64(10),
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to describe efs mount target for %s: %w", *fs.FileSystemId, err)
				}
				for _, mt := range mtRes.MountTargets {
					mtR := resource.New(ctx, resource.EfsMountTarget, mt.MountTargetId, mt.MountTargetId, mt)
					mtR.AddRelation(resource.EfsFileSystem, fs.FileSystemId, "")
					if mt.SubnetId != nil {
						mtR.AddRelation(resource.Ec2Subnet, mt.SubnetId, "")
					}
					rg.AddResource(mtR)
				}
				if mtRes.NextMarker == nil {
					break
				}
				mtMarker = mtRes.NextMarker
			}
			var apNextToken *string
			for {
				apRes, err := svc.DescribeAccessPointsRequest(&efs.DescribeAccessPointsInput{
					FileSystemId: fs.FileSystemId,
					MaxResults:   aws.Int64(100),
					NextToken:    apNextToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to describe efs access points for %s: %w", *fs.FileSystemId, err)
				}
				for _, ap := range apRes.AccessPoints {
					apR := resource.New(ctx, resource.EfsAccessPoint, ap.AccessPointId, ap.Name, ap)
					apR.AddRelation(resource.EfsFileSystem, fs.FileSystemId, "")
					rg.AddResource(apR)
				}
				if apRes.NextToken == nil {
					break
				}
				apNextToken = apRes.NextToken
			}
		}

		if res.NextMarker == nil {
			break
		}
		fsMarker = res.NextMarker
	}

	return rg, nil
}
