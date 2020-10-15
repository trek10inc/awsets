package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/trek10inc/awsets/option"
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

func (l AWSConfigDeliveryChannel) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := configservice.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()

	channels, err := svc.DescribeDeliveryChannels(cfg.Context, &configservice.DescribeDeliveryChannelsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list config delivery channels: %w", err)
	}
	for _, v := range channels.DeliveryChannels {
		r := resource.New(cfg, resource.ConfigDeliveryChannel, v.Name, v.Name, v)
		r.AddRelation(resource.S3Bucket, v.S3BucketName, "")
		rg.AddResource(r)
	}

	return rg, nil
}
