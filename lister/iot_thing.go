package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSIoTThing struct {
}

func init() {
	i := AWSIoTThing{}
	listers = append(listers, i)
}

func (l AWSIoTThing) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.IoTThing}
}

func (l AWSIoTThing) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := iot.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListThings(cfg.Context, &iot.ListThingsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list iot thing: %w", err)
		}
		for _, thing := range res.Things {
			r := resource.New(cfg, resource.IoTThing, thing.ThingName, thing.ThingName, thing)
			r.AddRelation(resource.IoTThingType, thing.ThingTypeName, "")
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
