package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
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

func (l AWSIoTThingType) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := iot.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListThingTypes(cfg.Context, &iot.ListThingTypesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list iot thing types: %w", err)
		}
		for _, thingType := range res.ThingTypes {
			ttArn := arn.ParseP(thingType.ThingTypeArn)
			r := resource.New(cfg, resource.IoTThingType, ttArn.ResourceId, thingType.ThingTypeName, thingType)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
