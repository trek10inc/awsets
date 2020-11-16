package lister

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := iam.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listIAMRolesOnce.Do(func() {
		outerErr = Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListRoles(ctx.Context, &iam.ListRolesInput{
				MaxItems: aws.Int32(100),
				Marker:   nt,
			})
			if err != nil {
				return nil, err
			}
			for _, role := range res.Roles {
				r := resource.NewGlobal(ctx, resource.IamRole, role.RoleName, role.RoleName, role)
				rg.AddResource(r)
			}
			return res.Marker, nil
		})
	})

	return rg, outerErr
}
