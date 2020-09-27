package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

var listUsersOnce sync.Once

type AWSIamUser struct {
}

func init() {
	i := AWSIamUser{}
	listers = append(listers, i)
}

func (l AWSIamUser) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.IamUser,
		resource.IamAccessKey,
	}
}

func (l AWSIamUser) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := iam.New(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error
	listUsersOnce.Do(func() {
		paginator := iam.NewListUsersPaginator(svc.ListUsersRequest(&iam.ListUsersInput{
			MaxItems: aws.Int64(100),
		}))
		for paginator.Next(ctx.Context) {
			page := paginator.CurrentPage()
			for _, user := range page.Users {
				r := resource.NewGlobal(ctx, resource.IamUser, user.UserId, user.UserName, user)

				akPaginator := iam.NewListAccessKeysPaginator(svc.ListAccessKeysRequest(&iam.ListAccessKeysInput{
					MaxItems: aws.Int64(100),
					UserName: user.UserName,
				}))
				for akPaginator.Next(ctx.Context) {
					akPage := akPaginator.CurrentPage()
					for _, ak := range akPage.AccessKeyMetadata {
						akR := resource.NewGlobal(ctx, resource.IamAccessKey, ak.AccessKeyId, ak.AccessKeyId, ak)
						akR.AddRelation(resource.IamUser, user.UserId, "")

						lastUsed, err := svc.GetAccessKeyLastUsedRequest(&iam.GetAccessKeyLastUsedInput{
							AccessKeyId: ak.AccessKeyId,
						}).Send(ctx.Context)
						if err != nil {
							outerErr = fmt.Errorf("failed to get lasted used time for access key %s: %w", *ak.AccessKeyId, err)
							return
						}
						akR.AddAttribute("LastUsed", lastUsed.AccessKeyLastUsed)
						rg.AddResource(akR)
					}
				}

				rg.AddResource(r)
			}
		}
		outerErr = paginator.Err()
	})

	return rg, outerErr
}
