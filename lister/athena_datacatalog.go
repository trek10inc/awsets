package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSAthenaDataCatalog struct {
}

func init() {
	i := AWSAthenaDataCatalog{}
	listers = append(listers, i)
}

func (l AWSAthenaDataCatalog) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.AthenaDataCatalog}
}

func (l AWSAthenaDataCatalog) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := athena.New(ctx.AWSCfg)

	rg := resource.NewGroup()

	req := svc.ListDataCatalogsRequest(&athena.ListDataCatalogsInput{
		MaxResults: aws.Int64(50),
	})

	paginator := athena.NewListDataCatalogsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, dcSummary := range page.DataCatalogsSummary {
			r := resource.New(ctx, resource.AthenaDataCatalog, dcSummary.CatalogName, dcSummary.CatalogName, dcSummary)

			dc, err := svc.GetDataCatalogRequest(&athena.GetDataCatalogInput{
				Name: dcSummary.CatalogName,
			}).Send(ctx.Context)
			if err != nil {
				//ctx.Logger.Errorf("failed to get data catalog %s of type %v: %v\n", *dcSummary.CatalogName, dcSummary.Type, err)
			} else if v := dc.DataCatalog; v != nil {
				r.AddAttribute("Description", v.Description)
				r.AddAttribute("Parameters", v.Parameters)
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
