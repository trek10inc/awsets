package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSWafRegionalRegexPatternSet struct {
}

func init() {
	i := AWSWafRegionalRegexPatternSet{}
	listers = append(listers, i)
}

func (l AWSWafRegionalRegexPatternSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafRegionalRegexPatternSet}
}

func (l AWSWafRegionalRegexPatternSet) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListRegexPatternSets(cfg.Context, &wafregional.ListRegexPatternSetsInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list regional regex pattern sets: %w", err)
		}
		for _, id := range res.RegexPatternSets {
			matchSet, err := svc.GetRegexPatternSet(cfg.Context, &wafregional.GetRegexPatternSetInput{
				RegexPatternSetId: id.RegexPatternSetId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get regex pattern set %s: %w", *id.RegexPatternSetId, err)
			}
			if v := matchSet.RegexPatternSet; v != nil {
				r := resource.New(cfg, resource.WafRegionalRegexPatternSet, v.RegexPatternSetId, v.Name, v)
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
