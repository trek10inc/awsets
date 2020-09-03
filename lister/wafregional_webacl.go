package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
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

func (l AWSWafRegionalWebAcl) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := wafregional.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextMarker *string
	for {
		res, err := svc.ListWebACLsRequest(&wafregional.ListWebACLsInput{
			Limit:      aws.Int64(100),
			NextMarker: nextMarker,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list webacls: %w", err)
		}
		for _, webaclId := range res.WebACLs {
			webacl, err := svc.GetWebACLRequest(&wafregional.GetWebACLInput{WebACLId: webaclId.WebACLId}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get webacl %s: %w", aws.StringValue(webaclId.WebACLId), err)
			}
			if v := webacl.WebACL; v != nil {
				//webaclArn := arn.ParseP(webacl.WebACL.WebACLArn)
				r := resource.New(ctx, resource.WafRegionalWebACL, v.WebACLId, v.Name, v)
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
