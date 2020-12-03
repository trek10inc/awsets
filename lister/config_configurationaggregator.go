package lister

import (
	"fmt"

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

	svc := configservice.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeConfigurationAggregators(ctx.Context, &configservice.DescribeConfigurationAggregatorsInput{
			Limit:     100,
			NextToken: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list config aggregators: %w", err)
		}
		for _, v := range res.ConfigurationAggregators {
			r := resource.New(ctx, resource.ConfigConfigurationAggregator, v.ConfigurationAggregatorName, v.ConfigurationAggregatorName, v)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
