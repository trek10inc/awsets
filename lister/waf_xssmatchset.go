package lister

import (
	"fmt"
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
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
	svc := waf.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	var outerErr error

	listWafXssMatchSetsOnce.Do(func() {
		var nextMarker *string
		for {
			res, err := svc.ListXssMatchSetsRequest(&waf.ListXssMatchSetsInput{
				Limit:      aws.Int64(100),
				NextMarker: nextMarker,
			}).Send(ctx.Context)
			if err != nil {
				outerErr = fmt.Errorf("failed to list xss match sets: %w", err)
				return
			}
			for _, id := range res.XssMatchSets {
				xssMatchSet, err := svc.GetXssMatchSetRequest(&waf.GetXssMatchSetInput{
					XssMatchSetId: id.XssMatchSetId,
				}).Send(ctx.Context)
				if err != nil {
					outerErr = fmt.Errorf("failed to get xss match stringset %s: %w", aws.StringValue(id.XssMatchSetId), err)
					return
				}
				if v := xssMatchSet.XssMatchSet; v != nil {
					r := resource.NewGlobal(ctx, resource.WafXssMatchSet, v.XssMatchSetId, v.Name, v)
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
