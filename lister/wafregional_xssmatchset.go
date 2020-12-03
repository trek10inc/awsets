package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSWafRegionalXssMatchSet struct {
}

func init() {
	i := AWSWafRegionalXssMatchSet{}
	listers = append(listers, i)
}

func (l AWSWafRegionalXssMatchSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafRegionalXssMatchSet}
}

func (l AWSWafRegionalXssMatchSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListXssMatchSets(ctx.Context, &wafregional.ListXssMatchSetsInput{
			Limit:      100,
			NextMarker: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list regional xss match sets: %w", err)
		}
		for _, id := range res.XssMatchSets {
			matchSet, err := svc.GetXssMatchSet(ctx.Context, &wafregional.GetXssMatchSetInput{
				XssMatchSetId: id.XssMatchSetId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get xss match set %s: %w", *id.XssMatchSetId, err)
			}
			if v := matchSet.XssMatchSet; v != nil {
				r := resource.New(ctx, resource.WafRegionalXssMatchSet, v.XssMatchSetId, v.Name, v)
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
