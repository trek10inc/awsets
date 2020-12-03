package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSServiceCatalogPortfolio struct {
}

func init() {
	i := AWSServiceCatalogPortfolio{}
	listers = append(listers, i)
}

func (l AWSServiceCatalogPortfolio) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ServiceCatalogPortfolio}
}

func (l AWSServiceCatalogPortfolio) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := servicecatalog.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListPortfolios(ctx.Context, &servicecatalog.ListPortfoliosInput{
			PageSize:  20,
			PageToken: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.PortfolioDetails {
			detail, err := svc.DescribePortfolio(ctx.Context, &servicecatalog.DescribePortfolioInput{
				Id: v.Id,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe service catalog portfolio %s: %w", *v.Id, err)
			}
			r := resource.New(ctx, resource.ServiceCatalogPortfolio, v.Id, v.DisplayName, detail)

			// Principals
			principals := make([]types.Principal, 0)
			err = Paginator(func(nt2 *string) (*string, error) {
				pRes, err := svc.ListPrincipalsForPortfolio(ctx.Context, &servicecatalog.ListPrincipalsForPortfolioInput{
					PortfolioId: v.Id,
					PageSize:    20,
					PageToken:   nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list principals for portfolio %s: %w", *v.Id, err)
				}
				if len(pRes.Principals) > 0 {
					principals = append(principals, pRes.Principals...)
				}
				return pRes.NextPageToken, nil
			})
			if err != nil {
				return nil, err
			}
			r.AddAttribute("Principals", principals)

			rg.AddResource(r)
		}
		return res.NextPageToken, nil
	})
	return rg, err
}
