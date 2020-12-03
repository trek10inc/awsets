package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/trek10inc/awsets/context"
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

func (l AWSConfigAggregationAuthorization) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := configservice.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeAggregationAuthorizations(ctx.Context, &configservice.DescribeAggregationAuthorizationsInput{
			Limit:     100,
			NextToken: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list config aggregation authorizations: %w", err)
		}
		for _, v := range res.AggregationAuthorizations {
			r := resource.New(ctx, resource.ConfigAggregationAuthorization, v.AuthorizedAccountId, v.AuthorizedAccountId, v)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
