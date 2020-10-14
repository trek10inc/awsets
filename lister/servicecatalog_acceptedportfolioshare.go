package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSServiceCatalogAcceptedPortfolioShare struct {
}

func init() {
	i := AWSServiceCatalogAcceptedPortfolioShare{}
	listers = append(listers, i)
}

func (l AWSServiceCatalogAcceptedPortfolioShare) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ServiceCatalogAcceptedPortfolioShare}
}

func (l AWSServiceCatalogAcceptedPortfolioShare) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := servicecatalog.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListAcceptedPortfolioShares(ctx.Context, &servicecatalog.ListAcceptedPortfolioSharesInput{
			PageSize:  aws.Int32(20),
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
			r := resource.New(ctx, resource.ServiceCatalogAcceptedPortfolioShare, v.Id, v.DisplayName, detail)
			rg.AddResource(r)
		}
		return res.NextPageToken, nil
	})
	return rg, err
}
