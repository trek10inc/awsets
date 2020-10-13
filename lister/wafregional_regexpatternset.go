package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
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

func (l AWSWafRegionalRegexPatternSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextMarker *string
	for {
		res, err := svc.ListRegexPatternSets(ctx.Context, &wafregional.ListRegexPatternSetsInput{
			Limit:      aws.Int32(100),
			NextMarker: nextMarker,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list regional regex pattern sets: %w", err)
		}
		for _, id := range res.RegexPatternSets {
			matchSet, err := svc.GetRegexPatternSet(ctx.Context, &wafregional.GetRegexPatternSetInput{
				RegexPatternSetId: id.RegexPatternSetId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get regex pattern set %s: %w", aws.StringValue(id.RegexPatternSetId), err)
			}
			if v := matchSet.RegexPatternSet; v != nil {
				r := resource.New(ctx, resource.WafRegionalRegexPatternSet, v.RegexPatternSetId, v.Name, v)
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
