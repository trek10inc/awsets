package lister

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := iam.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listInstanceProfilesOnce.Do(func() {
		outerErr = Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListInstanceProfiles(ctx.Context, &iam.ListInstanceProfilesInput{
				MaxItems: aws.Int32(100),
				Marker:   nt,
			})
			if err != nil {
				return nil, err
			}
			for _, profile := range res.InstanceProfiles {
				r := resource.NewGlobal(ctx, resource.IamInstanceProfile, profile.InstanceProfileId, profile.InstanceProfileName, profile)
				rg.AddResource(r)
			}
			return res.Marker, nil
		})
	})

	return rg, outerErr
}
