package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSServiceDiscoveryNamespace struct {
}

func init() {
	i := AWSServiceDiscoveryNamespace{}
	listers = append(listers, i)
}

func (l AWSServiceDiscoveryNamespace) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ServiceDiscoveryNamespace,
	}
}

func (l AWSServiceDiscoveryNamespace) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := servicediscovery.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListNamespaces(cfg.Context, &servicediscovery.ListNamespacesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, ns := range res.Namespaces {
			v, err := svc.GetNamespace(cfg.Context, &servicediscovery.GetNamespaceInput{
				Id: ns.Id,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe service discovery namespace %s: %w", *ns.Id, err)
			}
			r := resource.New(cfg, resource.ServiceDiscoveryNamespace, v.Namespace.Id, v.Namespace.Name, v.Namespace)
			if v.Namespace.Properties != nil {
				if v.Namespace.Properties.DnsProperties != nil {
					r.AddRelation(resource.Route53HostedZone, *v.Namespace.Properties.DnsProperties.HostedZoneId, "")
				}
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
