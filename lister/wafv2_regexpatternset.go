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

	rg, err := wafv2RegexPatternSetQuery(ctx, types.ScopeRegional)
	if err != nil {
		return nil, fmt.Errorf("failed to list ipsets: %w", err)
	}

	// Do global
	var outerErr error
	listWafv2RegexPatternSetOnce.Do(func() {
		ctxUsEast := ctx.Copy("us-east-1")
		rgNew, err := wafv2RegexPatternSetQuery(ctxUsEast, types.ScopeCloudfront)
		if err != nil {
			outerErr = fmt.Errorf("failed to list global regex pattern sets: %w", err)
		}
		rg.Merge(rgNew)
	})

	return rg, outerErr
}

func wafv2RegexPatternSetQuery(ctx context.AWSetsCtx, scope types.Scope) (*resource.Group, error) {
	svc := wafv2.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListRegexPatternSets(ctx.Context, &wafv2.ListRegexPatternSetsInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
			Scope:      scope,
		})
		if err != nil {
			return nil, err
		}
		for _, rpsId := range res.RegexPatternSets {
			rps, err := svc.GetRegexPatternSet(ctx.Context, &wafv2.GetRegexPatternSetInput{
				Id:    rpsId.Id,
				Name:  rpsId.Name,
				Scope: scope,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get regex pattern set %s for scope %v: %w", *rpsId.Id, scope, err)
			}
			if v := rps.RegexPatternSet; v != nil {
				var r resource.Resource
				if scope == types.ScopeCloudfront {
					r = resource.NewGlobal(ctx, resource.Wafv2RegexPatternSet, v.Id, v.Name, v)
				} else {
					r = resource.New(ctx, resource.Wafv2RegexPatternSet, v.Id, v.Name, v)
				}
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
