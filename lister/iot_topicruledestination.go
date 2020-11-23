package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSIoTTopicRuleDestination struct {
}

func init() {
	i := AWSIoTTopicRuleDestination{}
	listers = append(listers, i)
}

func (l AWSIoTTopicRuleDestination) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.IotTopicRuleDestination}
}

func (l AWSIoTTopicRuleDestination) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := iot.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListTopicRuleDestinations(ctx.Context, &iot.ListTopicRuleDestinationsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list iot topic rule destinations: %w", err)
		}
		for _, destination := range res.DestinationSummaries {
			dArn := arn.ParseP(destination.Arn)
			r := resource.New(ctx, resource.IotTopicRuleDestination, dArn.ResourceId, dArn.ResourceId, destination)

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
