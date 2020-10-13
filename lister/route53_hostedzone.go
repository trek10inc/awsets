package lister

import (
	"fmt"
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/route53"
)

var listRoute53HostedZonesOnce sync.Once

type AWSRoute53HostedZone struct {
}

func init() {
	i := AWSRoute53HostedZone{}
	listers = append(listers, i)
}

func (l AWSRoute53HostedZone) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Route53HostedZone, resource.Route53RecordSet}
}

func (l AWSRoute53HostedZone) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := route53.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listRoute53HostedZonesOnce.Do(func() {
		res, err := svc.ListHostedZones(ctx.Context, &route53.ListHostedZonesInput{})
		paginator := route53.NewListHostedZonesPaginator(req)
		for paginator.Next(ctx.Context) {
			page := paginator.CurrentPage()
			for _, hostedZone := range page.HostedZones {

				r := resource.NewGlobal(ctx, resource.Route53HostedZone, hostedZone.Id, hostedZone.Name, hostedZone)

				rsPaginator := route53.NewListResourceRecordSetsPaginator(svc.ListResourceRecordSets(ctx.Context, &route53.ListResourceRecordSetsInput{
					HostedZoneId: hostedZone.Id,
				}))
				for rsPaginator.Next(ctx.Context) {
					rsPage := rsPaginator.CurrentPage()
					for _, rs := range rsPage.ResourceRecordSets {
						rsRes := resource.NewGlobal(ctx, resource.Route53RecordSet, rs.Name, rs.Name, rs)
						rsRes.AddRelation(resource.Route53HostedZone, hostedZone.Id, "")
						rsRes.AddRelation(resource.Route53HealthCheck, rs.HealthCheckId, "")
						rg.AddResource(rsRes)
					}
				}
				err := rsPaginator.Err()
				if err != nil {
					outerErr = fmt.Errorf("failed to list record sets for hosted zone %s: %w", *hostedZone.Name, err)
					return
				}
				rg.AddResource(r)
			}
		}
		outerErr = paginator.Err()
	})

	return rg, outerErr
}
