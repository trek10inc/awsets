package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
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

func (l AWSKinesisStream) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := kinesis.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListStreams(cfg.Context, &kinesis.ListStreamsInput{
			Limit:                    aws.Int32(100),
			ExclusiveStartStreamName: nt,
		})
		if err != nil {
			return nil, err
		}
		var lastName *string
		for i, stream := range res.StreamNames {
			lastName = res.StreamNames[i]
			res, err := svc.DescribeStream(cfg.Context, &kinesis.DescribeStreamInput{
				Limit:      aws.Int32(100),
				StreamName: stream,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe kinesis streams %s: %w", *stream, err)
			}
			streamArn := arn.ParseP(res.StreamDescription.StreamARN)
			r := resource.New(cfg, resource.KinesisStream, streamArn.ResourceId, res.StreamDescription.StreamName, res.StreamDescription)
			rg.AddResource(r)
			// TODO the rest of this... relationships to shards and whatnot
		}
		return lastName, nil
	})
	return rg, err
}
