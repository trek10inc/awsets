package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

var listWafSizeConstraintSetsOnce sync.Once

type AWSWafSizeConstraintSet struct {
}

func init() {
	i := AWSWafSizeConstraintSet{}
	listers = append(listers, i)
}

func (l AWSWafSizeConstraintSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafSizeConstraintSet}
}

func (l AWSWafSizeConstraintSet) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := waf.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()

	var outerErr error

	listWafSizeConstraintSetsOnce.Do(func() {
		outerErr = Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListSizeConstraintSets(cfg.Context, &waf.ListSizeConstraintSetsInput{
				Limit:      aws.Int32(100),
				NextMarker: nt,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list size constraint sets: %w", err)
			}
			for _, id := range res.SizeConstraintSets {
				sizeConstraintSet, err := svc.GetSizeConstraintSet(cfg.Context, &waf.GetSizeConstraintSetInput{
					SizeConstraintSetId: id.SizeConstraintSetId,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get size constraint stringset %s: %w", *id.SizeConstraintSetId, err)
				}
				if v := sizeConstraintSet.SizeConstraintSet; v != nil {
					r := resource.NewGlobal(cfg, resource.WafSizeConstraintSet, v.SizeConstraintSetId, v.Name, v)
					rg.AddResource(r)
				}
			}
			return res.NextMarker, nil
		})
	})

	return rg, outerErr
}
