package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/option"
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

func (l AWSWafRegionalByteMatchSet) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListByteMatchSets(cfg.Context, &wafregional.ListByteMatchSetsInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list regional byte match sets: %w", err)
		}
		for _, id := range res.ByteMatchSets {
			matchSet, err := svc.GetByteMatchSet(cfg.Context, &wafregional.GetByteMatchSetInput{
				ByteMatchSetId: id.ByteMatchSetId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get byte match set %s: %w", *id.ByteMatchSetId, err)
			}
			if v := matchSet.ByteMatchSet; v != nil {
				r := resource.New(cfg, resource.WafRegionalByteMatchSet, v.ByteMatchSetId, v.Name, v)
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
