package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/trek10inc/awsets/option"
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

func (l AWSConfigConfigurationAggregator) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := configservice.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeConfigurationAggregators(cfg.Context, &configservice.DescribeConfigurationAggregatorsInput{
			Limit:     aws.Int32(100),
			NextToken: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list config aggregators: %w", err)
		}
		for _, v := range res.ConfigurationAggregators {
			r := resource.New(cfg, resource.ConfigConfigurationAggregator, v.ConfigurationAggregatorName, v.ConfigurationAggregatorName, v)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
