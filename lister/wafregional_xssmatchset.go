package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/option"
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

func (l AWSWafRegionalXssMatchSet) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListXssMatchSets(cfg.Context, &wafregional.ListXssMatchSetsInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list regional xss match sets: %w", err)
		}
		for _, id := range res.XssMatchSets {
			matchSet, err := svc.GetXssMatchSet(cfg.Context, &wafregional.GetXssMatchSetInput{
				XssMatchSetId: id.XssMatchSetId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get xss match set %s: %w", *id.XssMatchSetId, err)
			}
			if v := matchSet.XssMatchSet; v != nil {
				r := resource.New(cfg, resource.WafRegionalXssMatchSet, v.XssMatchSetId, v.Name, v)
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
