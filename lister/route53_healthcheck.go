package lister

import (
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/route53"
)

var listRoute53HealthChecksOnce sync.Once

type AWSRoute53HealthCheck struct {
}

func init() {
	i := AWSRoute53HealthCheck{}
	listers = append(listers, i)
}

func (l AWSRoute53HealthCheck) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Route53HealthCheck}
}

func (l AWSRoute53HealthCheck) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := route53.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listRoute53HealthChecksOnce.Do(func() {
		res, err := svc.ListHealthChecks(ctx.Context, &route53.ListHealthChecksInput{})
		paginator := route53.NewListHealthChecksPaginator(req)
		for paginator.Next(ctx.Context) {
			page := paginator.CurrentPage()
			for _, healthCheck := range page.HealthChecks {
				r := resource.NewGlobal(ctx, resource.Route53HealthCheck, healthCheck.Id, healthCheck.Id, healthCheck)
				rg.AddResource(r)
			}
		}
		outerErr = paginator.Err()
	})

	return rg, outerErr
}
