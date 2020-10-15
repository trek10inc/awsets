package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSSnsTopic struct {
}

func init() {
	i := AWSSnsTopic{}
	listers = append(listers, i)
}

func (l AWSSnsTopic) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.SnsTopic, resource.SnsSubscription}
}

func (l AWSSnsTopic) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := sns.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListTopics(cfg.Context, &sns.ListTopicsInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, topic := range res.Topics {
			topicArn := arn.ParseP(topic.TopicArn)
			r := resource.New(cfg, resource.SnsTopic, topicArn.ResourceId, "", topic)

			// Subscriptions
			err = Paginator(func(nt2 *string) (*string, error) {
				subs, err := svc.ListSubscriptionsByTopic(cfg.Context, &sns.ListSubscriptionsByTopicInput{
					TopicArn: topic.TopicArn,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get subscriptions for topic %s: %w", *topic.TopicArn, err)
				}
				for _, sub := range subs.Subscriptions {
					if arn.IsArnP(sub.SubscriptionArn) {
						subArn := arn.ParseP(sub.SubscriptionArn)
						subR := resource.New(cfg, resource.SnsSubscription, subArn.ResourceId, "", sub)
						subR.AddRelation(resource.SnsTopic, topicArn.ResourceId, "")
						rg.AddResource(subR)
					}
				}
				return subs.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Attributes
			res, err := svc.GetTopicAttributes(cfg.Context, &sns.GetTopicAttributesInput{TopicArn: topic.TopicArn})
			if err != nil {
				return nil, fmt.Errorf("failed to query topic attributes for %s: %w\n", *topic.TopicArn, err)
			}
			for k, v := range res.Attributes {
				r.AddAttribute(k, v)
			}

			// TODO: tags. policy?

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
