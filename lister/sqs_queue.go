package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/trek10inc/awsets/arn"
)

type AWSSqsQueue struct {
}

func init() {
	i := AWSSqsQueue{}
	listers = append(listers, i)
}

func (l AWSSqsQueue) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.SqsQueue}
}

func (l AWSSqsQueue) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := sqs.NewFromConfig(ctx.AWSCfg)

	res, err := svc.ListQueues(ctx.Context, &sqs.ListQueuesInput{
		MaxResults: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := sqs.NewListQueuesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, queue := range page.QueueUrls {
			qRes, err := svc.GetQueueAttributes(ctx.Context, &sqs.GetQueueAttributesInput{
				AttributeNames: []sqs.QueueAttributeName{sqs.QueueAttributeNameAll},
				QueueUrl:       aws.String(queue),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get queue attributes %s: %w", queue, err)
			}
			queueArn := arn.Parse(qRes.Attributes["QueueArn"])
			asMap := make(map[string]interface{})
			for k, v := range qRes.Attributes {
				asMap[k] = v
			}
			tagRes, err := svc.ListQueueTags(ctx.Context, &sqs.ListQueueTagsInput{
				QueueUrl: aws.String(queue),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get queue tags %s: %w", queue, err)
			}
			asMap["Tags"] = tagRes.Tags
			r := resource.New(ctx, resource.SqsQueue, queue, queueArn.ResourceId, asMap)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
