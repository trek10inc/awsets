package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSWafRegionalWebAcl struct {
}

func init() {
	i := AWSWafRegionalWebAcl{}
	listers = append(listers, i)
}

func (l AWSWafRegionalWebAcl) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafRegionalWebACL}
}

func (l AWSWafRegionalWebAcl) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListWebACLs(cfg.Context, &wafregional.ListWebACLsInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list webacls: %w", err)
		}
		for _, webaclId := range res.WebACLs {
			webacl, err := svc.GetWebACL(cfg.Context, &wafregional.GetWebACLInput{WebACLId: webaclId.WebACLId})
			if err != nil {
				return nil, fmt.Errorf("failed to get webacl %s: %w", *webaclId.WebACLId, err)
			}
			if v := webacl.WebACL; v != nil {
				//webaclArn := arn.ParseP(webacl.WebACL.WebACLArn)
				r := resource.New(cfg, resource.WafRegionalWebACL, v.WebACLId, v.Name, v)
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
