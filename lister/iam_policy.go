package lister

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := iam.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listPoliciesOnce.Do(func() {
		outerErr = Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListPolicies(ctx.Context, &iam.ListPoliciesInput{
				MaxItems: aws.Int32(100),
				Marker:   nt,
			})
			if err != nil {
				return nil, err
			}
			for _, policy := range res.Policies {
				r := resource.NewGlobal(ctx, resource.IamPolicy, policy.PolicyId, policy.PolicyName, policy)
				rg.AddResource(r)
			}
			return res.Marker, nil
		})
	})

	return rg, outerErr
}
