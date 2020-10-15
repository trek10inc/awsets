package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/option"
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

func (l AWSWafRegionalGeoMatchSet) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListGeoMatchSets(cfg.Context, &wafregional.ListGeoMatchSetsInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list regional geo match sets: %w", err)
		}
		for _, id := range res.GeoMatchSets {
			matchSet, err := svc.GetGeoMatchSet(cfg.Context, &wafregional.GetGeoMatchSetInput{
				GeoMatchSetId: id.GeoMatchSetId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get geo match set %s: %w", *id.GeoMatchSetId, err)
			}
			if v := matchSet.GeoMatchSet; v != nil {
				r := resource.New(cfg, resource.WafRegionalGeoMatchSet, v.GeoMatchSetId, v.Name, v)
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
