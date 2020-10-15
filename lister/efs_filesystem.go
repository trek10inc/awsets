package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
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

func (l AWSEfsFileSystems) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := efs.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeFileSystems(cfg.Context, &efs.DescribeFileSystemsInput{
			Marker:   nt,
			MaxItems: aws.Int32(10),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe efs file systems: %w", err)
		}
		for _, fs := range res.FileSystems {
			r := resource.New(cfg, resource.EfsFileSystem, fs.FileSystemId, fs.Name, fs)
			r.AddARNRelation(resource.KmsKey, fs.KmsKeyId)
			rg.AddResource(r)

			// Mount Targets
			err = Paginator(func(nt2 *string) (*string, error) {
				mounts, err := svc.DescribeMountTargets(cfg.Context, &efs.DescribeMountTargetsInput{
					FileSystemId: fs.FileSystemId,
					Marker:       nt2,
					MaxItems:     aws.Int32(10),
				})
				if err != nil {
					return nil, fmt.Errorf("failed to describe efs mount target for %s: %w", *fs.FileSystemId, err)
				}
				for _, mt := range mounts.MountTargets {
					mtR := resource.New(cfg, resource.EfsMountTarget, mt.MountTargetId, mt.MountTargetId, mt)
					mtR.AddRelation(resource.EfsFileSystem, fs.FileSystemId, "")
					if mt.SubnetId != nil {
						mtR.AddRelation(resource.Ec2Subnet, mt.SubnetId, "")
					}
					rg.AddResource(mtR)
				}
				return mounts.Marker, nil
			})
			if err != nil {
				return nil, err
			}

			// Access Points
			err = Paginator(func(nt2 *string) (*string, error) {
				points, err := svc.DescribeAccessPoints(cfg.Context, &efs.DescribeAccessPointsInput{
					FileSystemId: fs.FileSystemId,
					MaxResults:   aws.Int32(100),
					NextToken:    nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to describe efs access points for %s: %w", *fs.FileSystemId, err)
				}
				for _, ap := range points.AccessPoints {
					apR := resource.New(cfg, resource.EfsAccessPoint, ap.AccessPointId, ap.Name, ap)
					apR.AddRelation(resource.EfsFileSystem, fs.FileSystemId, "")
					rg.AddResource(apR)
				}
				return points.NextToken, nil
			})
			if err != nil {
				return nil, err
			}
		}
		return res.Marker, nil
	})
	return rg, err
}
