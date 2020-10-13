package lister

import (
	"fmt"
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
)

var listWafIpSetsOnce sync.Once

type AWSWafIpSet struct {
}

func init() {
	i := AWSWafIpSet{}
	listers = append(listers, i)
}

func (l AWSWafIpSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafIpSet}
}

func (l AWSWafIpSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := waf.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	var outerErr error

	listWafIpSetsOnce.Do(func() {
		var nextMarker *string
		for {
			res, err := svc.ListIPSets(ctx.Context, &waf.ListIPSetsInput{
				Limit:      aws.Int32(100),
				NextMarker: nextMarker,
			})
			if err != nil {
				outerErr = err
				return

			}
			for _, ipsetId := range res.IPSets {
				ipset, err := svc.GetIPSet(ctx.Context, &waf.GetIPSetInput{IPSetId: ipsetId.IPSetId})
				if err != nil {
					outerErr = fmt.Errorf("failed to get ipset %s: %w", aws.StringValue(ipsetId.IPSetId), err)
					return
				}
				if ipset.IPSet == nil {
					continue
				}
				r := resource.NewGlobal(ctx, resource.WafIpSet, ipset.IPSet.IPSetId, ipset.IPSet.Name, ipset.IPSet)
				rg.AddResource(r)
			}
			if res.NextMarker == nil {
				break
			}
			nextMarker = res.NextMarker
		}
	})

	return rg, outerErr
}
