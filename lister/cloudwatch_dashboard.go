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

	svc := cloudwatch.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListDashboards(ctx.Context, &cloudwatch.ListDashboardsInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list cloudwatch dashbards: %w", err)
		}
		for _, v := range res.DashboardEntries {
			dashboard, err := svc.GetDashboard(ctx.Context, &cloudwatch.GetDashboardInput{
				DashboardName: v.DashboardName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get dashboard %s: %w", *v.DashboardName, err)
			}
			r := resource.New(ctx, resource.CloudwatchDashboard, dashboard.DashboardName, dashboard.DashboardName, dashboard)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
