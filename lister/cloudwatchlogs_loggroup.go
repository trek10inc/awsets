package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSCloudwatchLogsGroups struct {
}

func init() {
	i := AWSCloudwatchLogsGroups{}
	listers = append(listers, i)
}

func (l AWSCloudwatchLogsGroups) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.LogGroup,
		resource.LogSubscriptionFilter,
		resource.LogMetricFilter,
		//resource.LogStream,
	}
}

func (l AWSCloudwatchLogsGroups) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloudwatchlogs.New(ctx.AWSCfg)

	req := svc.DescribeLogGroupsRequest(&cloudwatchlogs.DescribeLogGroupsInput{
		Limit: aws.Int64(50),
	})

	rg := resource.NewGroup()
	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.LogGroups {
			//groupArn := arn.ParseP(v.Arn)
			//r := resource.New(ctx, resourcetype.LogGroup, groupArn.ResourceId, aws.StringValue(v.LogGroupName), v) // TODO switch back to this after fixing ARN parsing
			r := resource.New(ctx, resource.LogGroup, v.LogGroupName, v.LogGroupName, v)
			if v.KmsKeyId != nil {
				kmsArn := arn.ParseP(v.KmsKeyId)
				r.AddRelation(resource.KmsKey, kmsArn.ResourceId, "")
			}
			rg.AddResource(r)

			subscriptionFilterReq := svc.DescribeSubscriptionFiltersRequest(&cloudwatchlogs.DescribeSubscriptionFiltersInput{
				Limit:        aws.Int64(50),
				LogGroupName: v.LogGroupName,
			})
			subscriptionFilterPaginator := cloudwatchlogs.NewDescribeSubscriptionFiltersPaginator(subscriptionFilterReq)
			for subscriptionFilterPaginator.Next(ctx.Context) {
				subFilterPage := subscriptionFilterPaginator.CurrentPage()
				for _, subFilter := range subFilterPage.SubscriptionFilters {
					subResource := resource.New(ctx, resource.LogSubscriptionFilter, subFilter.FilterName, subFilter.FilterName, subFilter)
					subResource.AddRelation(resource.LogGroup, v.LogGroupName, "")
					if subFilter.RoleArn != nil {
						roleArn := arn.ParseP(subFilter.RoleArn)
						subResource.AddRelation(resource.IamRole, roleArn.ResourceId, "")
					}
					rg.AddResource(subResource)
				}
			}

			metricFilterReq := svc.DescribeMetricFiltersRequest(&cloudwatchlogs.DescribeMetricFiltersInput{
				Limit:        aws.Int64(50),
				LogGroupName: v.LogGroupName,
			})
			metricFilterPaginator := cloudwatchlogs.NewDescribeMetricFiltersPaginator(metricFilterReq)
			for metricFilterPaginator.Next(ctx.Context) {
				metricFilterPage := metricFilterPaginator.CurrentPage()
				for _, metricFilter := range metricFilterPage.MetricFilters {
					mfResource := resource.New(ctx, resource.LogMetricFilter, metricFilter.FilterName, metricFilter.FilterName, metricFilter)
					mfResource.AddRelation(resource.LogGroup, v.LogGroupName, "")
					rg.AddResource(mfResource)
				}
			}

			/*
				// Querying log streams can take a _very_ long time, and doesn't provide a lot of value
				streamPaginator := cloudwatchlogs.NewDescribeLogStreamsPaginator(svc.DescribeLogStreamsRequest(&cloudwatchlogs.DescribeLogStreamsInput{
					Limit:        aws.Int64(50),
					LogGroupName: v.LogGroupName,
				}))
				for streamPaginator.Next(ctx.Context) {
					streamPage := streamPaginator.CurrentPage()
					for _, stream := range streamPage.LogStreams {
						streamArn := arn.ParseP(stream.Arn)
						streamResource := resource.New(ctx, resourcetype.LogStream, streamArn.ResourceId, aws.StringValue(stream.LogStreamName), stream)
						streamResource.AddRelation(resourcetype.LogGroup, groupArn.ResourceId, groupArn.ResourceVersion)
						rg.AddResource(streamResource)
					}
				}
				err := streamPaginator.Err()
				if err != nil {
					return rg, fmt.Errorf("failed to retrieve log streams for %s: %w", *v.LogGroupName, err)
				}
			*/
		}
	}
	err := paginator.Err()
	return rg, err
}
