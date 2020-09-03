package lister

import (
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

var listInstanceProfilesOnce sync.Once

type AWSIamInstanceProfile struct {
}

func init() {
	i := AWSIamInstanceProfile{}
	listers = append(listers, i)
}

func (l AWSIamInstanceProfile) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.IamInstanceProfile}
}

func (l AWSIamInstanceProfile) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := iam.New(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listInstanceProfilesOnce.Do(func() {
		req := svc.ListInstanceProfilesRequest(&iam.ListInstanceProfilesInput{
			MaxItems: aws.Int64(100),
		})

		paginator := iam.NewListInstanceProfilesPaginator(req)
		for paginator.Next(ctx.Context) {
			page := paginator.CurrentPage()
			for _, profile := range page.InstanceProfiles {
				r := resource.NewGlobal(ctx, resource.IamInstanceProfile, profile.InstanceProfileId, profile.InstanceProfileName, profile)
				rg.AddResource(r)
			}
		}
		outerErr = paginator.Err()
	})

	return rg, outerErr
}
