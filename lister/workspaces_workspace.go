package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/workspaces"

	"github.com/trek10inc/awsets/context"
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

func (l AWSWorkspacesWorkspace) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := workspaces.New(ctx.AWSCfg)

	req := svc.DescribeWorkspacesRequest(&workspaces.DescribeWorkspacesInput{})

	rg := resource.NewGroup()
	paginator := workspaces.NewDescribeWorkspacesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Workspaces {
			r := resource.New(ctx, resource.WorkspacesWorkspace, v.WorkspaceId, v.WorkspaceId, v)
			r.AddRelation(resource.Ec2Subnet, v.SubnetId, "")
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
