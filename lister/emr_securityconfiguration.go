package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/emr"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSEMRSecurityConfiguration struct {
}

func init() {
	i := AWSEMRSecurityConfiguration{}
	listers = append(listers, i)
}

func (l AWSEMRSecurityConfiguration) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.EmrSecurityConfiguration}
}

func (l AWSEMRSecurityConfiguration) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := emr.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListSecurityConfigurations(cfg.Context, &emr.ListSecurityConfigurationsInput{})
		if err != nil {
			return nil, err
		}
		for _, id := range res.SecurityConfigurations {
			v, err := svc.DescribeSecurityConfiguration(cfg.Context, &emr.DescribeSecurityConfigurationInput{
				Name: id.Name,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe security config %s: %w", *id.Name, err)
			}
			r := resource.New(cfg, resource.EmrSecurityConfiguration, v.Name, v.Name, v)
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
