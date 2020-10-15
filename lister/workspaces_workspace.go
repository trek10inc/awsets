package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSWorkspacesWorkspace struct {
}

func init() {
	i := AWSWorkspacesWorkspace{}
	listers = append(listers, i)
}

func (l AWSWorkspacesWorkspace) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WorkspacesWorkspace}
}

func (l AWSWorkspacesWorkspace) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := workspaces.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeWorkspaces(cfg.Context, &workspaces.DescribeWorkspacesInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.Workspaces {
			r := resource.New(cfg, resource.WorkspacesWorkspace, v.WorkspaceId, v.WorkspaceId, v)
			r.AddRelation(resource.Ec2Subnet, v.SubnetId, "")
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
