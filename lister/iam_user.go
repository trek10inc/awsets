package lister

import (
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

var listUsersOnce sync.Once

type AWSIamUser struct {
}

func init() {
	i := AWSIamUser{}
	listers = append(listers, i)
}

func (l AWSIamUser) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.IamUser}
}

func (l AWSIamUser) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := iam.New(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error
	listUsersOnce.Do(func() {
		req := svc.ListUsersRequest(&iam.ListUsersInput{
			MaxItems: aws.Int64(100),
		})

		paginator := iam.NewListUsersPaginator(req)
		for paginator.Next(ctx.Context) {
			page := paginator.CurrentPage()
			for _, user := range page.Users {
				r := resource.NewGlobal(ctx, resource.IamUser, user.UserId, user.UserName, user)
				rg.AddResource(r)
			}
		}
		outerErr = paginator.Err()
	})

	return rg, outerErr
}
