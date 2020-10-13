package lister

import (
	"fmt"
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/resource"
)

var listWafWebAclsOnce sync.Once

type AWSWafWebAcl struct {
}

func init() {
	i := AWSWafWebAcl{}
	listers = append(listers, i)
}

func (l AWSWafWebAcl) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafWebACL}
}

func (l AWSWafWebAcl) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := waf.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	var outerErr error

	listWafWebAclsOnce.Do(func() {
		var nextMarker *string
		for {
			res, err := svc.ListWebACLs(ctx.Context, &waf.ListWebACLsInput{
				Limit:      aws.Int32(100),
				NextMarker: nextMarker,
			})
			if err != nil {
				outerErr = fmt.Errorf("failed to list webacls: %w", err)
				return
			}
			for _, webaclId := range res.WebACLs {
				webacl, err := svc.GetWebACL(ctx.Context, &waf.GetWebACLInput{WebACLId: webaclId.WebACLId})
				if err != nil {
					outerErr = fmt.Errorf("failed to get webacl %s: %w", aws.StringValue(webaclId.WebACLId), err)
					return
				}
				if webacl.WebACL == nil {
					continue
				}
				webaclArn := arn.ParseP(webacl.WebACL.WebACLArn)
				r := resource.NewGlobal(ctx, resource.WafWebACL, webaclArn.ResourceId, webacl.WebACL.Name, webacl.WebACL)
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
