package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	"github.com/trek10inc/awsets/option"
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

func (l AWSFSxFileSystem) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := fsx.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeFileSystems(cfg.Context, &fsx.DescribeFileSystemsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.FileSystems {
			r := resource.New(cfg, resource.FSxFileSystem, v.FileSystemId, v.FileSystemId, v)
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
		return res.NextToken, nil
	})
	return rg, err
}
