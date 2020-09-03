package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
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

func (l AWSWafRegionalSizeConstraintSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := wafregional.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextMarker *string
	for {
		res, err := svc.ListSizeConstraintSetsRequest(&wafregional.ListSizeConstraintSetsInput{
			Limit:      aws.Int64(100),
			NextMarker: nextMarker,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list regional size constraint sets: %w", err)
		}
		for _, id := range res.SizeConstraintSets {
			matchSet, err := svc.GetSizeConstraintSetRequest(&wafregional.GetSizeConstraintSetInput{
				SizeConstraintSetId: id.SizeConstraintSetId,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get size constraint set %s: %w", aws.StringValue(id.SizeConstraintSetId), err)
			}
			if v := matchSet.SizeConstraintSet; v != nil {
				r := resource.New(ctx, resource.WafRegionalSizeConstraint, v.SizeConstraintSetId, v.Name, v)
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
