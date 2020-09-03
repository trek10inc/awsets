package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/resource"
)

type AWSWafRegionalIpSet struct {
}

func init() {
	i := AWSWafRegionalIpSet{}
	listers = append(listers, i)
}

func (l AWSWafRegionalIpSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafRegionalIpSet}
}

func (l AWSWafRegionalIpSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := wafregional.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextMarker *string
	for {
		res, err := svc.ListIPSetsRequest(&wafregional.ListIPSetsInput{
			Limit:      aws.Int64(100),
			NextMarker: nextMarker,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list ip sets: %w", err)
		}
		for _, id := range res.IPSets {
			ipset, err := svc.GetIPSetRequest(&wafregional.GetIPSetInput{
				IPSetId: id.IPSetId,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get ipset %s: %w", aws.StringValue(id.IPSetId), err)
			}
			if v := ipset.IPSet; v != nil {
				r := resource.New(ctx, resource.WafRegionalIpSet, v.IPSetId, v.Name, v)
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
