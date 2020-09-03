package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"
)

type AWSCloudwatchDashboard struct {
}

func init() {
	i := AWSCloudwatchDashboard{}
	listers = append(listers, i)
}

func (l AWSCloudwatchDashboard) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CloudwatchDashboard}
}

func (l AWSCloudwatchDashboard) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := cloudwatch.New(ctx.AWSCfg)

	rg := resource.NewGroup()

	var nextToken *string

	for {
		dashboards, err := svc.ListDashboardsRequest(&cloudwatch.ListDashboardsInput{
			NextToken: nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list cloudwatch dashbards: %w", err)
		}
		for _, v := range dashboards.DashboardEntries {
			dashboard, err := svc.GetDashboardRequest(&cloudwatch.GetDashboardInput{
				DashboardName: v.DashboardName,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get dashboard %s: %w", *v.DashboardName, err)
			}
			r := resource.New(ctx, resource.CloudwatchDashboard, dashboard.DashboardName, dashboard.DashboardName, dashboard)
			rg.AddResource(r)
		}
		if dashboards.NextToken == nil {
			break
		}
		nextToken = dashboards.NextToken
	}

	return rg, nil
}
