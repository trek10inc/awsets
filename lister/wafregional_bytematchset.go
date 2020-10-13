package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/resource"
)

type AWSWafRegionalByteMatchSet struct {
}

func init() {
	i := AWSWafRegionalByteMatchSet{}
	listers = append(listers, i)
}

func (l AWSWafRegionalByteMatchSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafRegionalByteMatchSet}
}

func (l AWSWafRegionalByteMatchSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextMarker *string
	for {
		res, err := svc.ListByteMatchSets(ctx.Context, &wafregional.ListByteMatchSetsInput{
			Limit:      aws.Int32(100),
			NextMarker: nextMarker,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list regional byte match sets: %w", err)
		}
		for _, id := range res.ByteMatchSets {
			matchSet, err := svc.GetByteMatchSet(ctx.Context, &wafregional.GetByteMatchSetInput{
				ByteMatchSetId: id.ByteMatchSetId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get byte match set %s: %w", aws.StringValue(id.ByteMatchSetId), err)
			}
			if v := matchSet.ByteMatchSet; v != nil {
				r := resource.New(ctx, resource.WafRegionalByteMatchSet, v.ByteMatchSetId, v.Name, v)
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
