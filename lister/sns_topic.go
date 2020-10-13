package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/trek10inc/awsets/arn"
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

func (l AWSSnsTopic) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := sns.NewFromConfig(ctx.AWSCfg)

	res, err := svc.ListTopics(ctx.Context, &sns.ListTopicsInput{})

	rg := resource.NewGroup()
	paginator := sns.NewListTopicsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, topic := range page.Topics {
			topicArn := arn.ParseP(topic.TopicArn)
			r := resource.New(ctx, resource.SnsTopic, topicArn.ResourceId, "", topic)

			subPag := sns.NewListSubscriptionsByTopicPaginator(svc.ListSubscriptionsByTopic(ctx.Context, &sns.ListSubscriptionsByTopicInput{
				TopicArn: topic.TopicArn,
			}))
			for subPag.Next(ctx.Context) {
				subs := subPag.CurrentPage()
				for _, sub := range subs.Subscriptions {
					if arn.IsArnP(sub.SubscriptionArn) {
						subArn := arn.ParseP(sub.SubscriptionArn)
						subR := resource.New(ctx, resource.SnsSubscription, subArn.ResourceId, "", sub)
						subR.AddRelation(resource.SnsTopic, topicArn.ResourceId, "")
						rg.AddResource(subR)
					}
				}
			}
			// TODO: tags. policy?

			res, err := svc.GetTopicAttributes(ctx.Context, &sns.GetTopicAttributesInput{TopicArn: topic.TopicArn})
			if err != nil {
				return nil, fmt.Errorf("failed to query topic attributes for %s: %w\n", aws.StringValue(topic.TopicArn), err)
			}
			for k, v := range res.Attributes {
				r.AddAttribute(k, v)
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
