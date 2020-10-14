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
		outerErr = Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListWebACLs(ctx.Context, &waf.ListWebACLsInput{
				Limit:      aws.Int32(100),
				NextMarker: nt,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list webacls: %w", err)
			}
			for _, webaclId := range res.WebACLs {
				webacl, err := svc.GetWebACL(ctx.Context, &waf.GetWebACLInput{WebACLId: webaclId.WebACLId})
				if err != nil {
					return nil, fmt.Errorf("failed to get webacl %s: %w", *webaclId.WebACLId, err)
				}
				if webacl.WebACL == nil {
					continue
				}
				webaclArn := arn.ParseP(webacl.WebACL.WebACLArn)
				r := resource.NewGlobal(ctx, resource.WafWebACL, webaclArn.ResourceId, webacl.WebACL.Name, webacl.WebACL)
				rg.AddResource(r)
			}
			return res.NextMarker, nil
		})
	})

	return rg, outerErr
}
