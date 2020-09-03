package lister

import (
	"fmt"
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
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

func (l AWSWafSizeConstraintSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := waf.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	var outerErr error

	listWafSizeConstraintSetsOnce.Do(func() {
		var nextMarker *string
		for {
			res, err := svc.ListSizeConstraintSetsRequest(&waf.ListSizeConstraintSetsInput{
				Limit:      aws.Int64(100),
				NextMarker: nextMarker,
			}).Send(ctx.Context)
			if err != nil {
				outerErr = fmt.Errorf("failed to list size constraint sets: %w", err)
				return
			}
			for _, id := range res.SizeConstraintSets {
				sizeConstraintSet, err := svc.GetSizeConstraintSetRequest(&waf.GetSizeConstraintSetInput{
					SizeConstraintSetId: id.SizeConstraintSetId,
				}).Send(ctx.Context)
				if err != nil {
					outerErr = fmt.Errorf("failed to get size constraint stringset %s: %w", aws.StringValue(id.SizeConstraintSetId), err)
					return
				}
				if v := sizeConstraintSet.SizeConstraintSet; v != nil {
					r := resource.NewGlobal(ctx, resource.WafSizeConstraintSet, v.SizeConstraintSetId, v.Name, v)
					rg.AddResource(r)
				}
			}
			if res.NextMarker == nil {
				break
			}
			nextMarker = res.NextMarker
		}
	})

	return rg, outerErr
}
