package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSSsmPatchBaseline struct {
}

func init() {
	i := AWSSsmPatchBaseline{}
	listers = append(listers, i)
}

func (l AWSSsmPatchBaseline) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.SsmPatchBaseline}
}

func (l AWSSsmPatchBaseline) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ssm.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribePatchBaselines(ctx.Context, &ssm.DescribePatchBaselinesInput{
			MaxResults: 50,
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, bl := range res.BaselineIdentities {
			v, err := svc.GetPatchBaseline(ctx.Context, &ssm.GetPatchBaselineInput{
				BaselineId: bl.BaselineId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get patch baseline %s: %w", *bl.BaselineId, err)
			}

			r := resource.New(ctx, resource.SsmPatchBaseline, v.BaselineId, v.Name, v)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
