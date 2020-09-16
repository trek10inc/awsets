package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/wafv2"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
)

var listWafv2RegexPatternSetOnce sync.Once

type AWSWafv2RegexPatternSet struct {
}

func init() {
	i := AWSWafv2RegexPatternSet{}
	listers = append(listers, i)
}

func (l AWSWafv2RegexPatternSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Wafv2RegexPatternSet}
}

func (l AWSWafv2RegexPatternSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	rg := resource.NewGroup()

	rg, err := wafv2RegexPatternSetQuery(ctx, wafv2.ScopeRegional)
	if err != nil {
		return rg, fmt.Errorf("failed to list ipsets: %w", err)
	}

	// Do global
	var outerErr error
	listWafv2RegexPatternSetOnce.Do(func() {
		ctxUsEast := ctx.Copy("us-east-1")
		rgNew, err := wafv2RegexPatternSetQuery(ctxUsEast, wafv2.ScopeCloudfront)
		if err != nil {
			outerErr = fmt.Errorf("failed to list global regex pattern sets: %w", err)
		}
		rg.Merge(rgNew)
	})

	return rg, outerErr
}

func wafv2RegexPatternSetQuery(ctx context.AWSetsCtx, scope wafv2.Scope) (*resource.Group, error) {
	svc := wafv2.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextMarker *string
	for {
		res, err := svc.ListRegexPatternSetsRequest(&wafv2.ListRegexPatternSetsInput{
			Limit:      aws.Int64(100),
			NextMarker: nextMarker,
			Scope:      scope,
		}).Send(ctx.Context)
		if err != nil {
			return rg, err
		}
		for _, rpsId := range res.RegexPatternSets {
			rps, err := svc.GetRegexPatternSetRequest(&wafv2.GetRegexPatternSetInput{
				Id:    rpsId.Id,
				Name:  rpsId.Name,
				Scope: scope,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get regex pattern set %s for scope %v: %w", *rpsId.Id, scope, err)
			}
			if v := rps.RegexPatternSet; v != nil {
				var r resource.Resource
				if scope == wafv2.ScopeCloudfront {
					r = resource.NewGlobal(ctx, resource.Wafv2RegexPatternSet, v.Id, v.Name, v)
				} else {
					r = resource.New(ctx, resource.Wafv2RegexPatternSet, v.Id, v.Name, v)
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
