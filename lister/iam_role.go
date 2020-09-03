package lister

import (
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

var listIAMRolesOnce sync.Once

type AWSIamRole struct {
}

func init() {
	i := AWSIamRole{}
	listers = append(listers, i)
}

func (l AWSIamRole) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.IamRole}
}

func (l AWSIamRole) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := iam.New(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listIAMRolesOnce.Do(func() {
		req := svc.ListRolesRequest(&iam.ListRolesInput{
			MaxItems: aws.Int64(100),
		})

		paginator := iam.NewListRolesPaginator(req)
		for paginator.Next(ctx.Context) {
			page := paginator.CurrentPage()
			for _, role := range page.Roles {
				r := resource.NewGlobal(ctx, resource.IamRole, role.RoleName, role.RoleName, role)
				rg.AddResource(r)
			}
		}
		outerErr = paginator.Err()
	})

	return rg, outerErr
}
