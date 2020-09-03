package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/emr"
	"github.com/trek10inc/awsets/context"
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

func (l AWSEMRSecurityConfiguration) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := emr.New(ctx.AWSCfg)

	req := svc.ListSecurityConfigurationsRequest(&emr.ListSecurityConfigurationsInput{})
	rg := resource.NewGroup()
	paginator := emr.NewListSecurityConfigurationsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, id := range page.SecurityConfigurations {
			v, err := svc.DescribeSecurityConfigurationRequest(&emr.DescribeSecurityConfigurationInput{
				Name: id.Name,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to describe security config %s: %w", *id.Name, err)
			}
			r := resource.New(ctx, resource.EmrSecurityConfiguration, v.Name, v.Name, v)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
