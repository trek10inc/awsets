package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSDynamoDBStreamStream struct {
}

func init() {
	i := AWSDynamoDBStreamStream{}
	listers = append(listers, i)
}

func (l AWSDynamoDBStreamStream) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.DynamoDbStreamStream,
	}
}

func (l AWSDynamoDBStreamStream) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := dynamodbstreams.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListStreams(ctx.Context, &dynamodbstreams.ListStreamsInput{
			ExclusiveStartStreamArn: nt,
			Limit:                   aws.Int32(100),
		})
		if err != nil {
			return nil, err
		}
		for _, stream := range res.Streams {
			v, err := svc.DescribeStream(ctx.Context, &dynamodbstreams.DescribeStreamInput{
				StreamArn: stream.StreamArn,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe ddb stream %s: %w", *stream.TableName, err)
			}
			streamArn := arn.ParseP(v.StreamDescription.StreamArn)
			r := resource.New(ctx, resource.DynamoDbStreamStream, streamArn.ResourceId, streamArn.ResourceId, v.StreamDescription)
			r.AddRelation(resource.DynamoDbTable, v.StreamDescription.TableName, "")
			rg.AddResource(r)
		}
		return res.LastEvaluatedStreamArn, nil
	})
	return rg, err
}
