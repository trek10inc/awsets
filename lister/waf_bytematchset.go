package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

var listWafByteMatchSetsOnce sync.Once

type AWSWafByteMatchSet struct {
}

func init() {
	i := AWSWafByteMatchSet{}
	listers = append(listers, i)
}

func (l AWSWafByteMatchSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafByteMatchSet}
}

func (l AWSWafByteMatchSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := waf.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	var outerErr error

	listWafByteMatchSetsOnce.Do(func() {
		outerErr = Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListByteMatchSets(ctx.Context, &waf.ListByteMatchSetsInput{
				Limit:      aws.Int32(100),
				NextMarker: nt,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list byte match sets: %w", err)
			}
			for _, id := range res.ByteMatchSets {
				byteMatchSet, err := svc.GetByteMatchSet(ctx.Context, &waf.GetByteMatchSetInput{
					ByteMatchSetId: id.ByteMatchSetId,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get byte match stringset %s: %w", *id.ByteMatchSetId, err)
				}
				if v := byteMatchSet.ByteMatchSet; v != nil {
					r := resource.NewGlobal(ctx, resource.WafByteMatchSet, v.ByteMatchSetId, v.Name, v)
					rg.AddResource(r)
				}
			}
			return res.NextMarker, nil
		})
	})
	return rg, outerErr
}
