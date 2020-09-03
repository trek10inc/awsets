package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/configservice"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"
)

type AWSConfigConfigurationAggregator struct {
}

func init() {
	i := AWSConfigConfigurationAggregator{}
	listers = append(listers, i)
}

func (l AWSConfigConfigurationAggregator) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ConfigConfigurationAggregator}
}

func (l AWSConfigConfigurationAggregator) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := configservice.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextToken *string

	for {
		aggregators, err := svc.DescribeConfigurationAggregatorsRequest(&configservice.DescribeConfigurationAggregatorsInput{
			Limit:     aws.Int64(100),
			NextToken: nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list config aggregators: %w", err)
		}
		for _, v := range aggregators.ConfigurationAggregators {
			r := resource.New(ctx, resource.ConfigConfigurationAggregator, v.ConfigurationAggregatorName, v.ConfigurationAggregatorName, v)
			rg.AddResource(r)
		}
		if aggregators.NextToken == nil {
			break
		}
		nextToken = aggregators.NextToken
	}

	return rg, nil
}
