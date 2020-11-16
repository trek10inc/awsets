package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ses/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSSESConfigurationSet struct {
}

func init() {
	i := AWSSESConfigurationSet{}
	listers = append(listers, i)
}

func (l AWSSESConfigurationSet) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.SesConfigurationSet,
		resource.SesConfigurationSetEventDestination,
	}
}

func (l AWSSESConfigurationSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ses.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListConfigurationSets(ctx.Context, &ses.ListConfigurationSetsInput{
			MaxItems:  aws.Int32(10),
			NextToken: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, cs := range res.ConfigurationSets {

			v, err := svc.DescribeConfigurationSet(ctx.Context, &ses.DescribeConfigurationSetInput{
				ConfigurationSetName: cs.Name,
				ConfigurationSetAttributeNames: []types.ConfigurationSetAttribute{
					types.ConfigurationSetAttributeEventDestinations,
					types.ConfigurationSetAttributeTrackingOptions,
					types.ConfigurationSetAttributeDeliveryOptions,
					types.ConfigurationSetAttributeReputationOptions,
				},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get ses configuration set %s: %w", *cs.Name, err)
			}
			r := resource.New(ctx, resource.SesConfigurationSet, v.ConfigurationSet.Name, v.ConfigurationSet.Name, v)
			if v.EventDestinations != nil {
				for _, ed := range v.EventDestinations {
					edr := resource.New(ctx, resource.SesConfigurationSetEventDestination, ed.Name, ed.Name, ed)
					edr.AddRelation(resource.SesConfigurationSet, v.ConfigurationSet.Name, "")
					if ed.KinesisFirehoseDestination != nil {
						edr.AddARNRelation(resource.FirehoseDeliveryStream, ed.KinesisFirehoseDestination.DeliveryStreamARN)
						edr.AddARNRelation(resource.IamRole, ed.KinesisFirehoseDestination.IAMRoleARN)
					}
					if ed.SNSDestination != nil {
						edr.AddARNRelation(resource.SnsTopic, ed.SNSDestination.TopicARN)
					}
				}
			}

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
