package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/trek10inc/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/service/iot"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSIoTThingType struct {
}

func init() {
	i := AWSIoTThingType{}
	listers = append(listers, i)
}

func (l AWSIoTThingType) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.IoTThingType}
}

func (l AWSIoTThingType) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := iot.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		thingTypes, err := svc.ListThingTypes(ctx.Context, &iot.ListThingTypesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list iot thing types: %w", err)
		}
		for _, thingType := range thingTypes.ThingTypes {
			ttArn := arn.ParseP(thingType.ThingTypeArn)
			r := resource.New(ctx, resource.IoTThingType, ttArn.ResourceId, thingType.ThingTypeName, thingType)
			rg.AddResource(r)
		}
		if thingTypes.NextToken == nil {
			break
		}
		nextToken = thingTypes.NextToken
	}
	return rg, nil
}
