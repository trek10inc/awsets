package lister

import (
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
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
	svc := iam.New(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listGroupsOnce.Do(func() {
		req := svc.ListGroupsRequest(&iam.ListGroupsInput{
			MaxItems: aws.Int64(100),
		})

		paginator := iam.NewListGroupsPaginator(req)
		for paginator.Next(ctx.Context) {
			page := paginator.CurrentPage()
			for _, group := range page.Groups {
				r := resource.NewGlobal(ctx, resource.IamGroup, group.GroupId, group.GroupName, group)
				rg.AddResource(r)
			}
		}
		outerErr = paginator.Err()
	})

	return rg, outerErr
}
