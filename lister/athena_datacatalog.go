package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/trek10inc/awsets/option"
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

func (l AWSAthenaDataCatalog) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := athena.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListDataCatalogs(cfg.Context, &athena.ListDataCatalogsInput{
			MaxResults: aws.Int32(50),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, dcSummary := range res.DataCatalogsSummary {
			r := resource.New(cfg, resource.AthenaDataCatalog, dcSummary.CatalogName, dcSummary.CatalogName, dcSummary)

			dc, err := svc.GetDataCatalog(cfg.Context, &athena.GetDataCatalogInput{
				Name: dcSummary.CatalogName,
			})
			if err != nil {
				//cfg.SendStatus(option.StatusLogError, fmt.Sprintf("failed to get data catalog %s of type %v: %v\n", *dcSummary.CatalogName, dcSummary.Type, err))
			} else if v := dc.DataCatalog; v != nil {
				r.AddAttribute("Description", v.Description)
				r.AddAttribute("Parameters", v.Parameters)
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})

	return rg, err
}
