package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSCloudwatchAlarm struct {
}

func init() {
	i := AWSCloudwatchAlarm{}
	listers = append(listers, i)
}

func (l AWSCloudwatchAlarm) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CloudwatchAlarm}
}

func (l AWSCloudwatchAlarm) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := cloudwatch.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeAlarms(cfg.Context, &cloudwatch.DescribeAlarmsInput{
			MaxRecords: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, alarm := range res.CompositeAlarms {
			alarmArn := arn.ParseP(alarm.AlarmArn)
			r := resource.New(cfg, resource.CloudwatchAlarm, alarmArn.ResourceId, alarm.AlarmName, alarm)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
