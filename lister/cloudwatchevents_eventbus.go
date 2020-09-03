package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchevents"
	"github.com/trek10inc/awsets/resource"
)

type AWSCloudwatchEventsBus struct {
}

func init() {
	i := AWSCloudwatchEventsBus{}
	listers = append(listers, i)
}

func (l AWSCloudwatchEventsBus) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.EventsBus}
}

func (l AWSCloudwatchEventsBus) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := cloudwatchevents.New(ctx.AWSCfg)

	rg := resource.NewGroup()

	var nextToken *string

	for {
		buses, err := svc.ListEventBusesRequest(&cloudwatchevents.ListEventBusesInput{
			Limit:     aws.Int64(100),
			NextToken: nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list cloudwatch event buses: %w", err)
		}
		for _, bus := range buses.EventBuses {
			r := resource.New(ctx, resource.EventsBus, bus.Name, bus.Name, bus)
			rg.AddResource(r)
		}
		if buses.NextToken == nil {
			break
		}
		nextToken = buses.NextToken
	}
	return rg, nil
}
