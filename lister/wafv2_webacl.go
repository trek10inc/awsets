package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

var listWafv2WebACLOnce sync.Once

type AWSWafv2WebACL struct {
}

func init() {
	i := AWSWafv2WebACL{}
	listers = append(listers, i)
}

func (l AWSWafv2WebACL) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Wafv2WebACL}
}

func (l AWSWafv2WebACL) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	rg := resource.NewGroup()

	rg, err := wafv2WebACLQuery(ctx, types.ScopeRegional)
	if err != nil {
		return nil, fmt.Errorf("failed to list web ACLs: %w", err)
	}

	// Do global
	var outerErr error
	listWafv2WebACLOnce.Do(func() {
		ctxUsEast := ctx.Copy("us-east-1")
		rgNew, err := wafv2WebACLQuery(*ctxUsEast, types.ScopeCloudfront)
		if err != nil {
			outerErr = fmt.Errorf("failed to list global web ACLs: %w", err)
		}
		rg.Merge(rgNew)
	})

	return rg, outerErr
}

func wafv2WebACLQuery(ctx context.AWSetsCtx, scope types.Scope) (*resource.Group, error) {
	svc := wafv2.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListWebACLs(ctx.Context, &wafv2.ListWebACLsInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
			Scope:      scope,
		})
		if err != nil {
			return nil, err
		}
		for _, waId := range res.WebACLs {
			wa, err := svc.GetWebACL(ctx.Context, &wafv2.GetWebACLInput{
				Id:    waId.Id,
				Name:  waId.Name,
				Scope: scope,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get web ACL %s for scope %v: %w", *waId.Id, scope, err)
			}
			if v := wa.WebACL; v != nil {
				var r resource.Resource
				if scope == types.ScopeCloudfront {
					r = resource.NewGlobal(ctx, resource.Wafv2WebACL, v.Id, v.Name, v)
				} else {
					r = resource.New(ctx, resource.Wafv2WebACL, v.Id, v.Name, v)

					allResources := make([]string, 0)
					for _, t := range []types.ResourceType{
						types.ResourceTypeApiGateway,
						types.ResourceTypeApplicationLoadBalancer,
					} {
						resources, err := svc.ListResourcesForWebACL(ctx.Context, &wafv2.ListResourcesForWebACLInput{
							WebACLArn:    v.ARN,
							ResourceType: t,
						})
						if err != nil {
							return nil, fmt.Errorf("failed to get resources for web acl %s: %w", *v.Id, err)
						}
						allResources = append(allResources, resources.ResourceArns...)
						for _, res := range resources.ResourceArns {
							if t == types.ResourceTypeApiGateway {
								r.AddARNRelation(resource.ApiGatewayRestApi, res)
							} else if t == types.ResourceTypeApplicationLoadBalancer {
								r.AddARNRelation(resource.ElbV2LoadBalancer, res)
							}
						}
					}
					if len(allResources) > 0 {
						r.AddARNRelation("Resources", allResources)
					}
				}

				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
