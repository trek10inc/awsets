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
	svc := iam.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error
	listUsersOnce.Do(func() {
		outerErr = Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListUsers(ctx.Context, &iam.ListUsersInput{
				MaxItems: aws.Int32(100),
				Marker:   nt,
			})
			if err != nil {
				return nil, err
			}
			for _, user := range res.Users {
				r := resource.NewGlobal(ctx, resource.IamUser, user.UserId, user.UserName, user)

				// Access Keys
				err = Paginator(func(nt2 *string) (*string, error) {
					keys, err := svc.ListAccessKeys(ctx.Context, &iam.ListAccessKeysInput{
						MaxItems: aws.Int32(100),
						UserName: user.UserName,
						Marker:   nt2,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to get access keys for user %s: %w", *user.UserName, err)
					}
					for _, ak := range keys.AccessKeyMetadata {
						akR := resource.NewGlobal(ctx, resource.IamAccessKey, ak.AccessKeyId, ak.AccessKeyId, ak)
						akR.AddRelation(resource.IamUser, user.UserId, "")

						lastUsed, err := svc.GetAccessKeyLastUsed(ctx.Context, &iam.GetAccessKeyLastUsedInput{
							AccessKeyId: ak.AccessKeyId,
						})
						if err != nil {
							return nil, fmt.Errorf("failed to get lasted used time for access key %s: %w", *ak.AccessKeyId, err)
						}
						akR.AddAttribute("LastUsed", lastUsed.AccessKeyLastUsed)
						rg.AddResource(akR)
					}
					return res.Marker, nil
				})
				if err != nil {
					return nil, err
				}
				rg.AddResource(r)
			}
			return res.Marker, nil
		})
	})

	return rg, outerErr
}
