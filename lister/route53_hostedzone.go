package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
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

func (l AWSRoute53HostedZone) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := route53.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listRoute53HostedZonesOnce.Do(func() {
		outerErr = Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListHostedZones(cfg.Context, &route53.ListHostedZonesInput{
				Marker: nt,
			})
			if err != nil {
				return nil, err
			}
			for _, hostedZone := range res.HostedZones {

				r := resource.NewGlobal(cfg, resource.Route53HostedZone, hostedZone.Id, hostedZone.Name, hostedZone)

				// Record Sets
				err = Paginator(func(nt2 *string) (*string, error) {
					sets, err := svc.ListResourceRecordSets(cfg.Context, &route53.ListResourceRecordSetsInput{
						HostedZoneId:          hostedZone.Id,
						StartRecordIdentifier: nt,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to list record sets for hosted zone %s: %w", *hostedZone.Name, err)
					}
					for _, rs := range sets.ResourceRecordSets {
						rsRes := resource.NewGlobal(cfg, resource.Route53RecordSet, rs.Name, rs.Name, rs)
						rsRes.AddRelation(resource.Route53HostedZone, hostedZone.Id, "")
						rsRes.AddRelation(resource.Route53HealthCheck, rs.HealthCheckId, "")
						rg.AddResource(rsRes)
					}
					return sets.NextRecordIdentifier, nil
				})
				if err != nil {
					return nil, err
				}

				rg.AddResource(r)
			}
			return res.Marker, nil
		})
	})

	return rg, outerErr
}
