package lister

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

var listGroupsOnce sync.Once

type AWSIamGroup struct {
}

func init() {
	i := AWSIamGroup{}
	listers = append(listers, i)
}

func (l AWSIamGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.IamGroup}
}

func (l AWSIamGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := iam.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listGroupsOnce.Do(func() {
		outerErr = Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListGroups(ctx.Context, &iam.ListGroupsInput{
				MaxItems: aws.Int32(100),
				Marker:   nt,
			})
			if err != nil {
				return nil, err
			}
			for _, group := range res.Groups {
				r := resource.NewGlobal(ctx, resource.IamGroup, group.GroupId, group.GroupName, group)
				rg.AddResource(r)
			}
			return res.Marker, nil
		})
	})

	return rg, outerErr
}
