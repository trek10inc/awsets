package lister

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSSESReceiptFilter struct {
}

func init() {
	i := AWSSESReceiptFilter{}
	listers = append(listers, i)
}

func (l AWSSESReceiptFilter) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.SesReceiptFilter,
	}
}

func (l AWSSESReceiptFilter) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ses.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()

	filters, err := svc.ListReceiptFilters(cfg.Context, &ses.ListReceiptFiltersInput{})
	if err != nil {
		if strings.Contains(err.Error(), "Unavailable Operation") {
			// If SES isn't available in a region, returns Unavailable Operation error
			return rg, nil
		}
		return rg, err
	}
	for _, v := range filters.Filters {
		r := resource.New(cfg, resource.SesReceiptFilter, v.Name, v.Name, v)
		rg.AddResource(r)
	}
	return rg, err
}
