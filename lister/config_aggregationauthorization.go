package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"

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

	svc := configservice.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextToken *string

	for {
		aggregations, err := svc.DescribeAggregationAuthorizationsRequest(&configservice.DescribeAggregationAuthorizationsInput{
			Limit:     aws.Int64(100),
			NextToken: nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list config aggregation authorizations: %w", err)
		}
		for _, v := range aggregations.AggregationAuthorizations {
			r := resource.New(ctx, resource.ConfigAggregationAuthorization, v.AuthorizedAccountId, v.AuthorizedAccountId, v)
			rg.AddResource(r)
		}
		if aggregations.NextToken == nil {
			break
		}
		nextToken = aggregations.NextToken
	}

	return rg, nil
}
