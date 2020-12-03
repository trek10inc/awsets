package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSAmazonMQBroker struct {
}

func init() {
	i := AWSAmazonMQBroker{}
	listers = append(listers, i)
}

func (l AWSAmazonMQBroker) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.AmazonMQBroker}
}

func (l AWSAmazonMQBroker) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := mq.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListBrokers(ctx.Context, &mq.ListBrokersInput{
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list mq brokers: %w", err)
		}
		for _, broker := range res.BrokerSummaries {

			v, err := svc.DescribeBroker(ctx.Context, &mq.DescribeBrokerInput{
				BrokerId: broker.BrokerId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe broker %s: %w", *broker.BrokerId, err)
			}
			r := resource.New(ctx, resource.AmazonMQBroker, v.BrokerId, v.BrokerName, v)

			for _, sg := range v.SecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, sg, "")
			}
			for _, sn := range v.SubnetIds {
				r.AddRelation(resource.Ec2Subnet, sn, "")
			}
			if conf := v.Configurations; conf != nil {
				r.AddRelation(resource.AmazonMQBrokerConfiguration, conf.Current.Id, fmt.Sprintf("%d", conf.Current.Revision))
				for _, c := range conf.History {
					r.AddRelation(resource.AmazonMQBrokerConfiguration, c.Id, fmt.Sprintf("%d", c.Revision))
				}
			}

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
