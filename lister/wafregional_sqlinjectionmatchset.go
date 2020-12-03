package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSWafRegionalSqlInjectionMatchSet struct {
}

func init() {
	i := AWSWafRegionalSqlInjectionMatchSet{}
	listers = append(listers, i)
}

func (l AWSWafRegionalSqlInjectionMatchSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafRegionalSqlInjectionMatchSet}
}

func (l AWSWafRegionalSqlInjectionMatchSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListSqlInjectionMatchSets(ctx.Context, &wafregional.ListSqlInjectionMatchSetsInput{
			Limit:      100,
			NextMarker: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list regional sql injection match sets: %w", err)
		}
		for _, id := range res.SqlInjectionMatchSets {
			matchSet, err := svc.GetSqlInjectionMatchSet(ctx.Context, &wafregional.GetSqlInjectionMatchSetInput{
				SqlInjectionMatchSetId: id.SqlInjectionMatchSetId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get sql injection match set %s: %w", *id.SqlInjectionMatchSetId, err)
			}
			if v := matchSet.SqlInjectionMatchSet; v != nil {
				r := resource.New(ctx, resource.WafRegionalSqlInjectionMatchSet, v.SqlInjectionMatchSetId, v.Name, v)
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
