package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/resource"
)

type AWSWafRegionalGeoMatchSet struct {
}

func init() {
	i := AWSWafRegionalGeoMatchSet{}
	listers = append(listers, i)
}

func (l AWSWafRegionalGeoMatchSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafRegionalGeoMatchSet}
}

func (l AWSWafRegionalGeoMatchSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextMarker *string
	for {
		res, err := svc.ListGeoMatchSets(ctx.Context, &wafregional.ListGeoMatchSetsInput{
			Limit:      aws.Int32(100),
			NextMarker: nextMarker,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list regional geo match sets: %w", err)
		}
		for _, id := range res.GeoMatchSets {
			matchSet, err := svc.GetGeoMatchSet(ctx.Context, &wafregional.GetGeoMatchSetInput{
				GeoMatchSetId: id.GeoMatchSetId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get geo match set %s: %w", aws.StringValue(id.GeoMatchSetId), err)
			}
			if v := matchSet.GeoMatchSet; v != nil {
				r := resource.New(ctx, resource.WafRegionalGeoMatchSet, v.GeoMatchSetId, v.Name, v)
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
