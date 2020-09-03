package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSFSxFileSystem struct {
}

func init() {
	i := AWSFSxFileSystem{}
	listers = append(listers, i)
}

func (l AWSFSxFileSystem) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.FSxFileSystem}
}

func (l AWSFSxFileSystem) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := fsx.New(ctx.AWSCfg)

	req := svc.DescribeFileSystemsRequest(&fsx.DescribeFileSystemsInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := fsx.NewDescribeFileSystemsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.FileSystems {
			r := resource.New(ctx, resource.FSxFileSystem, v.FileSystemId, v.FileSystemId, v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			for _, sn := range v.SubnetIds {
				r.AddRelation(resource.Ec2Subnet, sn, "")
			}
			for _, eni := range v.NetworkInterfaceIds {
				r.AddRelation(resource.Ec2NetworkInterface, eni, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
