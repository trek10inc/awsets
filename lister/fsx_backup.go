package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSFSxBackup struct {
}

func init() {
	i := AWSFSxBackup{}
	listers = append(listers, i)
}

func (l AWSFSxBackup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.FSxBackup}
}

func (l AWSFSxBackup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := fsx.NewFromConfig(ctx.AWSCfg)

	res, err := svc.DescribeBackups(ctx.Context, &fsx.DescribeBackupsInput{
		MaxResults: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := fsx.NewDescribeBackupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Backups {
			r := resource.New(ctx, resource.FSxBackup, v.BackupId, v.BackupId, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			if v.FileSystem != nil {
				r.AddRelation(resource.FSxFileSystem, v.FileSystem.FileSystemId, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
