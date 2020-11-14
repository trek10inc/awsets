package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

var listWafXssMatchSetsOnce sync.Once

type AWSWafXssMatchSet struct {
}

func init() {
	i := AWSWafXssMatchSet{}
	listers = append(listers, i)
}

func (l AWSWafXssMatchSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafXssMatchSet}
}

func (l AWSWafXssMatchSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := waf.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	var outerErr error

	listWafXssMatchSetsOnce.Do(func() {
		outerErr = Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListXssMatchSets(ctx.Context, &waf.ListXssMatchSetsInput{
				Limit:      aws.Int32(100),
				NextMarker: nt,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list xss match sets: %w", err)
			}
			for _, id := range res.XssMatchSets {
				xssMatchSet, err := svc.GetXssMatchSet(ctx.Context, &waf.GetXssMatchSetInput{
					XssMatchSetId: id.XssMatchSetId,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get xss match stringset %s: %w", *id.XssMatchSetId, err)
				}
				if v := xssMatchSet.XssMatchSet; v != nil {
					r := resource.NewGlobal(ctx, resource.WafXssMatchSet, v.XssMatchSetId, v.Name, v)
					rg.AddResource(r)
				}
			}
			return res.NextMarker, nil
		})
	})
	return rg, outerErr
}
