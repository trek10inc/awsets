package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := firehose.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListDeliveryStreams(ctx.Context, &firehose.ListDeliveryStreamsInput{
			ExclusiveStartDeliveryStreamName: nt,
			Limit:                            aws.Int32(10),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list delivery streams: %w", err)
		}
		var lastNameRead *string
		for _, ds := range res.DeliveryStreamNames {
			lastNameRead = ds
			describeRes, err := svc.DescribeDeliveryStream(ctx.Context, &firehose.DescribeDeliveryStreamInput{
				DeliveryStreamName:          ds,
				ExclusiveStartDestinationId: nil, // TODO actually page through these?
				Limit:                       aws.Int32(20),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe delivery stream %s: %w", *ds, err)
			}
			if *describeRes.DeliveryStreamDescription.HasMoreDestinations {
				ctx.Logger.Errorf("need to page through destinations for %s", ds)
			}
			dsArn := arn.ParseP(describeRes.DeliveryStreamDescription.DeliveryStreamARN)
			r := resource.New(ctx, resource.FirehoseDeliveryStream, dsArn.ResourceId, describeRes.DeliveryStreamDescription.DeliveryStreamName, describeRes.DeliveryStreamDescription)
			rg.AddResource(r)
		}
		return lastNameRead, nil
	})
	return rg, err
}
