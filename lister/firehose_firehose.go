package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/trek10inc/awsets/arn"
)

type AWSFirehoseStream struct {
}

func init() {
	i := AWSFirehoseStream{}
	listers = append(listers, i)
}

func (l AWSFirehoseStream) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.FirehoseDeliveryStream}
}

func (l AWSFirehoseStream) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := firehose.New(ctx.AWSCfg)

	rg := resource.NewGroup()

	var nextName *string
	for {
		deliveryStreams, err := svc.ListDeliveryStreamsRequest(&firehose.ListDeliveryStreamsInput{
			ExclusiveStartDeliveryStreamName: nextName,
			Limit:                            aws.Int64(10),
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list delivery streams: %w", err)
		}
		for _, ds := range deliveryStreams.DeliveryStreamNames {
			nextName = &ds
			describeRes, err := svc.DescribeDeliveryStreamRequest(&firehose.DescribeDeliveryStreamInput{
				DeliveryStreamName:          &ds,
				ExclusiveStartDestinationId: nil, // TODO actually page through these?
				Limit:                       aws.Int64(20),
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to describe delivery stream %s: %w", ds, err)
			}
			if *describeRes.DeliveryStreamDescription.HasMoreDestinations {
				ctx.Logger.Errorf("need to page through destinations for %s", ds)
			}
			dsArn := arn.ParseP(describeRes.DeliveryStreamDescription.DeliveryStreamARN)
			r := resource.New(ctx, resource.FirehoseDeliveryStream, dsArn.ResourceId, aws.StringValue(describeRes.DeliveryStreamDescription.DeliveryStreamName), describeRes.DeliveryStreamDescription)
			rg.AddResource(r)
		}
		if !*deliveryStreams.HasMoreDeliveryStreams {
			break
		}
	}
	return rg, nil
}
