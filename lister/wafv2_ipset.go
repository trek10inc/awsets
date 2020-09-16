package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
)

var listWafv2IpSetsOnce sync.Once

type AWSWafv2IpSet struct {
}

func init() {
	i := AWSWafv2IpSet{}
	listers = append(listers, i)
}

func (l AWSWafv2IpSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Wafv2IpSet}
}

func (l AWSWafv2IpSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	rg := resource.NewGroup()

	rg, err := wafv2IpsetQuery(ctx, wafv2.ScopeRegional)
	if err != nil {
		return rg, fmt.Errorf("failed to list ipsets: %w", err)
	}

	// Do global
	var outerErr error
	listWafv2IpSetsOnce.Do(func() {
		ctxUsEast := ctx.Copy("us-east-1")
		rgNew, err := wafv2IpsetQuery(ctxUsEast, wafv2.ScopeCloudfront)
		if err != nil {
			outerErr = fmt.Errorf("failed to list global ipsets in region %s: %w", ctxUsEast.Region(), err)
		}
		rg.Merge(rgNew)
	})

	return rg, outerErr
}

func wafv2IpsetQuery(ctx context.AWSetsCtx, scope wafv2.Scope) (*resource.Group, error) {
	svc := wafv2.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextMarker *string
	for {
		res, err := svc.ListIPSetsRequest(&wafv2.ListIPSetsInput{
			Limit:      aws.Int64(100),
			NextMarker: nextMarker,
			Scope:      scope,
		}).Send(ctx.Context)
		if err != nil {
			return rg, err
		}
		for _, ipsetId := range res.IPSets {
			ipset, err := svc.GetIPSetRequest(&wafv2.GetIPSetInput{
				Id:    ipsetId.Id,
				Name:  ipsetId.Name,
				Scope: scope,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get ipset %s for scope %v: %w", *ipsetId.Id, scope, err)
			}
			if v := ipset.IPSet; v != nil {
				var r resource.Resource
				if scope == wafv2.ScopeCloudfront {
					r = resource.NewGlobal(ctx, resource.Wafv2IpSet, v.Id, v.Name, v)
				} else {
					r = resource.New(ctx, resource.Wafv2IpSet, v.Id, v.Name, v)
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
