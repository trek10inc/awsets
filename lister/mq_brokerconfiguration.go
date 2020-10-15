package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSAmazonMQBrokerConfiguration struct {
}

func init() {
	i := AWSAmazonMQBrokerConfiguration{}
	listers = append(listers, i)
}

func (l AWSAmazonMQBrokerConfiguration) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.AmazonMQBrokerConfiguration}
}

func (l AWSAmazonMQBrokerConfiguration) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := mq.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListConfigurations(cfg.Context, &mq.ListConfigurationsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list mq broker configurations: %w", err)
		}
		for _, v := range res.Configurations {
			r := resource.New(cfg, resource.AmazonMQBrokerConfiguration, v.Id, v.Name, v)

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
