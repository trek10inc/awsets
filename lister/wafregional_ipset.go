package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/option"
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

func (l AWSWafRegionalIpSet) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListIPSets(cfg.Context, &wafregional.ListIPSetsInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list ip sets: %w", err)
		}
		for _, id := range res.IPSets {
			ipset, err := svc.GetIPSet(cfg.Context, &wafregional.GetIPSetInput{
				IPSetId: id.IPSetId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get ipset %s: %w", *id.IPSetId, err)
			}
			if v := ipset.IPSet; v != nil {
				r := resource.New(cfg, resource.WafRegionalIpSet, v.IPSetId, v.Name, v)
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
