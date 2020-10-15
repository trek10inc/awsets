package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSConfigAggregationAuthorization struct {
}

func init() {
	i := AWSConfigAggregationAuthorization{}
	listers = append(listers, i)
}

func (l AWSConfigAggregationAuthorization) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ConfigAggregationAuthorization}
}

func (l AWSConfigAggregationAuthorization) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := configservice.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeAggregationAuthorizations(cfg.Context, &configservice.DescribeAggregationAuthorizationsInput{
			Limit:     aws.Int32(100),
			NextToken: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list config aggregation authorizations: %w", err)
		}
		for _, v := range res.AggregationAuthorizations {
			r := resource.New(cfg, resource.ConfigAggregationAuthorization, v.AuthorizedAccountId, v.AuthorizedAccountId, v)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
