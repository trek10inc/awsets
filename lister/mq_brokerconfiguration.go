package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/trek10inc/awsets/context"
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

func (l AWSAmazonMQBrokerConfiguration) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := mq.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListConfigurations(ctx.Context, &mq.ListConfigurationsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list mq broker configurations: %w", err)
		}
		for _, v := range res.Configurations {
			r := resource.New(ctx, resource.AmazonMQBrokerConfiguration, v.Id, v.Name, v)

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
