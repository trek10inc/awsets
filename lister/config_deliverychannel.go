package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/configservice"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"
)

type AWSConfigDeliveryChannel struct {
}

func init() {
	i := AWSConfigDeliveryChannel{}
	listers = append(listers, i)
}

func (l AWSConfigDeliveryChannel) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ConfigDeliveryChannel}
}

func (l AWSConfigDeliveryChannel) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := configservice.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	channels, err := svc.DescribeDeliveryChannelsRequest(&configservice.DescribeDeliveryChannelsInput{}).Send(ctx.Context)
	if err != nil {
		return rg, fmt.Errorf("failed to list config delivery channels: %w", err)
	}
	for _, v := range channels.DeliveryChannels {
		r := resource.New(ctx, resource.ConfigDeliveryChannel, v.Name, v.Name, v)
		r.AddRelation(resource.S3Bucket, v.S3BucketName, "")
		rg.AddResource(r)
	}

	return rg, nil
}
