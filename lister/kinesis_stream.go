package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/trek10inc/awsets/arn"
)

type AWSKinesisStream struct {
}

func init() {
	i := AWSKinesisStream{}
	listers = append(listers, i)
}

func (l AWSKinesisStream) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.KinesisStream}
}

func (l AWSKinesisStream) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := kinesis.New(ctx.AWSCfg)
	req := svc.ListStreamsRequest(&kinesis.ListStreamsInput{
		Limit: aws.Int64(100),
	})
	paginator := kinesis.NewListStreamsPaginator(req)
	rg := resource.NewGroup()
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, stream := range page.StreamNames {
			res, err := svc.DescribeStreamRequest(&kinesis.DescribeStreamInput{
				Limit:      aws.Int64(100),
				StreamName: aws.String(stream),
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to describe kinesis streams %s: %w", stream, err)
			}
			streamArn := arn.ParseP(res.StreamDescription.StreamARN)
			r := resource.New(ctx, resource.KinesisStream, streamArn.ResourceId, res.StreamDescription.StreamName, res.StreamDescription)
			rg.AddResource(r)
			// TODO the rest of this... relationships to shards and whatnot
		}
	}
	err := paginator.Err()
	return rg, err
}
