package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListQueues(ctx.Context, &sqs.ListQueuesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, queue := range res.QueueUrls {
			qRes, err := svc.GetQueueAttributes(ctx.Context, &sqs.GetQueueAttributesInput{
				AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
				QueueUrl:       &queue,
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
				QueueUrl: &queue,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get queue tags %s: %w", queue, err)
			}
			asMap["Tags"] = tagRes.Tags
			r := resource.New(ctx, resource.SqsQueue, queue, queueArn.ResourceId, asMap)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
