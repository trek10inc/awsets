package lister

import (
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

var listPoliciesOnce sync.Once

type AWSIamPolicy struct {
}

func init() {
	i := AWSIamPolicy{}
	listers = append(listers, i)
}

func (l AWSIamPolicy) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.IamPolicy}
}

func (l AWSIamPolicy) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := iam.New(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listPoliciesOnce.Do(func() {
		req := svc.ListPoliciesRequest(&iam.ListPoliciesInput{
			MaxItems: aws.Int64(100),
		})

		paginator := iam.NewListPoliciesPaginator(req)
		for paginator.Next(ctx.Context) {
			page := paginator.CurrentPage()
			for _, policy := range page.Policies {
				r := resource.NewGlobal(ctx, resource.IamPolicy, policy.PolicyId, policy.PolicyName, policy)
				rg.AddResource(r)
			}
		}
		outerErr = paginator.Err()
	})

	return rg, outerErr
}
