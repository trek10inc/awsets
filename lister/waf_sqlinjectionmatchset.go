package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/waf"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

var listWafSqlInjectionMatchSetsOnce sync.Once

type AWSWafSqlInjectionMatchSet struct {
}

func init() {
	i := AWSWafSqlInjectionMatchSet{}
	listers = append(listers, i)
}

func (l AWSWafSqlInjectionMatchSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafSqlInjectionMatchSet}
}

func (l AWSWafSqlInjectionMatchSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := waf.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	var outerErr error

	listWafSqlInjectionMatchSetsOnce.Do(func() {
		outerErr = Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListSqlInjectionMatchSets(ctx.Context, &waf.ListSqlInjectionMatchSetsInput{
				Limit:      100,
				NextMarker: nt,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list sql injection match sets: %w", err)
			}
			for _, id := range res.SqlInjectionMatchSets {
				matchSet, err := svc.GetSqlInjectionMatchSet(ctx.Context, &waf.GetSqlInjectionMatchSetInput{
					SqlInjectionMatchSetId: id.SqlInjectionMatchSetId,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get sql injection match stringset %s: %w", *id.SqlInjectionMatchSetId, err)
				}
				if v := matchSet.SqlInjectionMatchSet; v != nil {
					r := resource.NewGlobal(ctx, resource.WafSqlInjectionMatchSet, v.SqlInjectionMatchSetId, v.Name, v)
					rg.AddResource(r)
				}
			}
			return res.NextMarker, nil
		})
	})

	return rg, outerErr
}
