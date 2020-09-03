package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"

	"github.com/trek10inc/awsets/context"

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

func (l AWSCloudwatchAlarm) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := cloudwatch.New(ctx.AWSCfg)

	req := svc.DescribeAlarmsRequest(&cloudwatch.DescribeAlarmsInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := cloudwatch.NewDescribeAlarmsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, alarm := range page.CompositeAlarms {
			alarmArn := arn.ParseP(alarm.AlarmArn)
			r := resource.New(ctx, resource.CloudwatchAlarm, alarmArn.ResourceId, alarm.AlarmName, alarm)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
