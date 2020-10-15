package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSWafRegionalSizeConstraintSet struct {
}

func init() {
	i := AWSWafRegionalSizeConstraintSet{}
	listers = append(listers, i)
}

func (l AWSWafRegionalSizeConstraintSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafRegionalSizeConstraint}
}

func (l AWSWafRegionalSizeConstraintSet) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListSizeConstraintSets(cfg.Context, &wafregional.ListSizeConstraintSetsInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list regional size constraint sets: %w", err)
		}
		for _, id := range res.SizeConstraintSets {
			matchSet, err := svc.GetSizeConstraintSet(cfg.Context, &wafregional.GetSizeConstraintSetInput{
				SizeConstraintSetId: id.SizeConstraintSetId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get size constraint set %s: %w", *id.SizeConstraintSetId, err)
			}
			if v := matchSet.SizeConstraintSet; v != nil {
				r := resource.New(cfg, resource.WafRegionalSizeConstraint, v.SizeConstraintSetId, v.Name, v)
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
