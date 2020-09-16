package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
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

	rg, err := wafv2WebACLQuery(ctx, wafv2.ScopeRegional)
	if err != nil {
		return rg, fmt.Errorf("failed to list web ACLs: %w", err)
	}

	// Do global
	var outerErr error
	listWafv2WebACLOnce.Do(func() {
		ctxUsEast := ctx.Copy("us-east-1")
		rgNew, err := wafv2WebACLQuery(ctxUsEast, wafv2.ScopeCloudfront)
		if err != nil {
			outerErr = fmt.Errorf("failed to list global web ACLs: %w", err)
		}
		rg.Merge(rgNew)
	})

	return rg, outerErr
}

func wafv2WebACLQuery(ctx context.AWSetsCtx, scope wafv2.Scope) (*resource.Group, error) {
	svc := wafv2.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextMarker *string
	for {
		res, err := svc.ListWebACLsRequest(&wafv2.ListWebACLsInput{
			Limit:      aws.Int64(100),
			NextMarker: nextMarker,
			Scope:      scope,
		}).Send(ctx.Context)
		if err != nil {
			return rg, err
		}
		for _, waId := range res.WebACLs {
			wa, err := svc.GetWebACLRequest(&wafv2.GetWebACLInput{
				Id:    waId.Id,
				Name:  waId.Name,
				Scope: scope,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get web ACL %s for scope %v: %w", *waId.Id, scope, err)
			}
			if v := wa.WebACL; v != nil {
				var r resource.Resource
				if scope == wafv2.ScopeCloudfront {
					r = resource.NewGlobal(ctx, resource.Wafv2WebACL, v.Id, v.Name, v)
				} else {
					r = resource.New(ctx, resource.Wafv2WebACL, v.Id, v.Name, v)
				}
				rg.AddResource(r)
			}
		}
		if res.NextMarker == nil {
			break
		}
		nextMarker = res.NextMarker
	}
	return rg, nil
}
