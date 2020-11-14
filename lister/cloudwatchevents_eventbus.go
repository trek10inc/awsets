package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchevents"
	"github.com/trek10inc/awsets/context"
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

	svc := cloudwatchevents.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListEventBuses(ctx.Context, &cloudwatchevents.ListEventBusesInput{
			Limit:     aws.Int32(100),
			NextToken: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list cloudwatch event buses: %w", err)
		}
		for _, bus := range res.EventBuses {
			r := resource.New(ctx, resource.EventsBus, bus.Name, bus.Name, bus)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
