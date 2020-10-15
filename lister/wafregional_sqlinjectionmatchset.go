package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/option"
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

func (l AWSWafRegionalSqlInjectionMatchSet) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListSqlInjectionMatchSets(cfg.Context, &wafregional.ListSqlInjectionMatchSetsInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list regional sql injection match sets: %w", err)
		}
		for _, id := range res.SqlInjectionMatchSets {
			matchSet, err := svc.GetSqlInjectionMatchSet(cfg.Context, &wafregional.GetSqlInjectionMatchSetInput{
				SqlInjectionMatchSetId: id.SqlInjectionMatchSetId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get sql injection match set %s: %w", *id.SqlInjectionMatchSetId, err)
			}
			if v := matchSet.SqlInjectionMatchSet; v != nil {
				r := resource.New(cfg, resource.WafRegionalSqlInjectionMatchSet, v.SqlInjectionMatchSetId, v.Name, v)
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
