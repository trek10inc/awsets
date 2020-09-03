package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
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
	svc := wafregional.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextMarker *string
	for {
		res, err := svc.ListSqlInjectionMatchSetsRequest(&wafregional.ListSqlInjectionMatchSetsInput{
			Limit:      aws.Int64(100),
			NextMarker: nextMarker,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list regional sql injection match sets: %w", err)
		}
		for _, id := range res.SqlInjectionMatchSets {
			matchSet, err := svc.GetSqlInjectionMatchSetRequest(&wafregional.GetSqlInjectionMatchSetInput{
				SqlInjectionMatchSetId: id.SqlInjectionMatchSetId,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get sql injection match set %s: %w", aws.StringValue(id.SqlInjectionMatchSetId), err)
			}
			if v := matchSet.SqlInjectionMatchSet; v != nil {
				r := resource.New(ctx, resource.WafRegionalSqlInjectionMatchSet, v.SqlInjectionMatchSetId, v.Name, v)
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
