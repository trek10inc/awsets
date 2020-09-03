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
	svc := servicecatalog.New(ctx.AWSCfg)

	req := svc.ListAcceptedPortfolioSharesRequest(&servicecatalog.ListAcceptedPortfolioSharesInput{
		PageSize: aws.Int64(20),
	})
	rg := resource.NewGroup()
	paginator := servicecatalog.NewListAcceptedPortfolioSharesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.PortfolioDetails {
			detail, err := svc.DescribePortfolioRequest(&servicecatalog.DescribePortfolioInput{
				Id: v.Id,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to describe service catalog portfolio %s: %w", *v.Id, err)
			}
			r := resource.New(ctx, resource.ServiceCatalogAcceptedPortfolioShare, v.Id, v.DisplayName, detail.DescribePortfolioOutput)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
