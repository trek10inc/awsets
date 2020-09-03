package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
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
	svc := servicecatalog.New(ctx.AWSCfg)

	req := svc.ListPortfoliosRequest(&servicecatalog.ListPortfoliosInput{
		PageSize: aws.Int64(20),
	})

	rg := resource.NewGroup()
	paginator := servicecatalog.NewListPortfoliosPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.PortfolioDetails {
			detail, err := svc.DescribePortfolioRequest(&servicecatalog.DescribePortfolioInput{
				Id: v.Id,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to describe service catalog portfolio %s: %w", *v.Id, err)
			}
			r := resource.New(ctx, resource.ServiceCatalogPortfolio, v.Id, v.DisplayName, detail.DescribePortfolioOutput)

			principalPaginator := servicecatalog.NewListPrincipalsForPortfolioPaginator(svc.ListPrincipalsForPortfolioRequest(&servicecatalog.ListPrincipalsForPortfolioInput{
				PortfolioId: v.Id,
				PageSize:    aws.Int64(20),
			}))
			principals := make([]servicecatalog.Principal, 0)
			for principalPaginator.Next(ctx.Context) {
				principalPage := principalPaginator.CurrentPage()
				if len(principalPage.Principals) > 0 {
					principals = append(principals, principalPage.Principals...)
				}
			}
			if err = principalPaginator.Err(); err != nil {
				return rg, fmt.Errorf("failed to list principals for portfolio %s: %w", *v.Id, err)
			}
			r.AddAttribute("Principals", principals)

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
