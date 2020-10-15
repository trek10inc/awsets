package lister

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
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

func (l AWSRoute53HealthCheck) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := route53.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	var outerErr error

	listRoute53HealthChecksOnce.Do(func() {
		outerErr = Paginator(func(nt *string) (*string, error) {
			res, err := svc.ListHealthChecks(cfg.Context, &route53.ListHealthChecksInput{
				Marker: nt,
			})
			if err != nil {
				return nil, err
			}
			for _, healthCheck := range res.HealthChecks {
				r := resource.NewGlobal(cfg, resource.Route53HealthCheck, healthCheck.Id, healthCheck.Id, healthCheck)
				rg.AddResource(r)
			}
			return res.Marker, nil
		})
	})

	return rg, outerErr
}
