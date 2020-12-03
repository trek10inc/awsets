package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSWafRegionalWebAcl struct {
}

func init() {
	i := AWSWafRegionalWebAcl{}
	listers = append(listers, i)
}

func (l AWSWafRegionalWebAcl) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafRegionalWebACL}
}

func (l AWSWafRegionalWebAcl) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListWebACLs(ctx.Context, &wafregional.ListWebACLsInput{
			Limit:      100,
			NextMarker: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list webacls: %w", err)
		}
		for _, webaclId := range res.WebACLs {
			webacl, err := svc.GetWebACL(ctx.Context, &wafregional.GetWebACLInput{WebACLId: webaclId.WebACLId})
			if err != nil {
				return nil, fmt.Errorf("failed to get webacl %s: %w", *webaclId.WebACLId, err)
			}
			if v := webacl.WebACL; v != nil {
				//webaclArn := arn.ParseP(webacl.WebACL.WebACLArn)
				r := resource.New(ctx, resource.WafRegionalWebACL, v.WebACLId, v.Name, v)

				allResources := make([]string, 0)
				for _, t := range []types.ResourceType{
					types.ResourceTypeApiGateway,
					types.ResourceTypeApplicationLoadBalancer,
				} {
					resources, err := svc.ListResourcesForWebACL(ctx.Context, &wafregional.ListResourcesForWebACLInput{
						WebACLId:     v.WebACLId,
						ResourceType: t,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to get resources for web acl %s: %w", *v.WebACLId, err)
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

				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
