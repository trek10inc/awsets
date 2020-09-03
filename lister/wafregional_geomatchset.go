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
	svc := wafregional.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextMarker *string
	for {
		res, err := svc.ListGeoMatchSetsRequest(&wafregional.ListGeoMatchSetsInput{
			Limit:      aws.Int64(100),
			NextMarker: nextMarker,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list regional geo match sets: %w", err)
		}
		for _, id := range res.GeoMatchSets {
			matchSet, err := svc.GetGeoMatchSetRequest(&wafregional.GetGeoMatchSetInput{
				GeoMatchSetId: id.GeoMatchSetId,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get geo match set %s: %w", aws.StringValue(id.GeoMatchSetId), err)
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
