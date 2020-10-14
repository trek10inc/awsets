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

	rg, err := wafv2IpsetQuery(ctx, types.ScopeRegional)
	if err != nil {
		return nil, fmt.Errorf("failed to list ipsets: %w", err)
	}

	// Do global
	var outerErr error
	listWafv2IpSetsOnce.Do(func() {
		ctxUsEast := ctx.Copy("us-east-1")
		rgNew, err := wafv2IpsetQuery(ctxUsEast, types.ScopeCloudfront)
		if err != nil {
			outerErr = fmt.Errorf("failed to list global ipsets in region %s: %w", ctxUsEast.Region(), err)
		}
		rg.Merge(rgNew)
	})
	return rg, outerErr
}

func wafv2IpsetQuery(ctx context.AWSetsCtx, scope types.Scope) (*resource.Group, error) {
	svc := wafv2.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListIPSets(ctx.Context, &wafv2.ListIPSetsInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
			Scope:      scope,
		})
		if err != nil {
			return nil, err
		}
		for _, ipsetId := range res.IPSets {
			ipset, err := svc.GetIPSet(ctx.Context, &wafv2.GetIPSetInput{
				Id:    ipsetId.Id,
				Name:  ipsetId.Name,
				Scope: scope,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get ipset %s for scope %v: %w", *ipsetId.Id, scope, err)
			}
			if v := ipset.IPSet; v != nil {
				var r resource.Resource
				if scope == types.ScopeCloudfront {
					r = resource.NewGlobal(ctx, resource.Wafv2IpSet, v.Id, v.Name, v)
				} else {
					r = resource.New(ctx, resource.Wafv2IpSet, v.Id, v.Name, v)
				}
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
