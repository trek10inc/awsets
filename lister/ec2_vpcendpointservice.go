package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2VpcEndpointService struct {
}

func init() {
	i := AWSEc2VpcEndpointService{}
	listers = append(listers, i)
}

func (l AWSEc2VpcEndpointService) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.Ec2VpcEndpointService,
	}
}

func (l AWSEc2VpcEndpointService) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeVpcEndpointServices(ctx.Context, &ec2.DescribeVpcEndpointServicesInput{
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.ServiceDetails {
			if v.ServiceId == nil {
				// some Amazon owned vpc service endpoints have null IDs
				continue
			}
			r := resource.New(ctx, resource.Ec2VpcEndpointService, v.ServiceId, v.ServiceName, v)

			configs := make([]types.ServiceConfiguration, 0)
			err = Paginator(func(nt2 *string) (*string, error) {
				scs, err := svc.DescribeVpcEndpointServiceConfigurations(ctx.Context, &ec2.DescribeVpcEndpointServiceConfigurationsInput{
					MaxResults: 100,
					NextToken:  nt2,
					ServiceIds: []string{*v.ServiceId},
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get vpc endpoint service configs for %s: %w", *v.ServiceId, err)
				}
				configs = append(configs, scs.ServiceConfigurations...)

				return scs.NextToken, nil
			})
			if err != nil {
				return nil, err
			}
			if len(configs) > 0 {
				r.AddAttribute("Configurations", configs)
			}

			principals := make([]types.AllowedPrincipal, 0)
			err = Paginator(func(nt2 *string) (*string, error) {
				perms, err := svc.DescribeVpcEndpointServicePermissions(ctx.Context, &ec2.DescribeVpcEndpointServicePermissionsInput{
					MaxResults: 100,
					NextToken:  nt2,
					ServiceId:  v.ServiceId,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get vpc endpoint service permissions for %s: %w", *v.ServiceId, err)
				}
				principals = append(principals, perms.AllowedPrincipals...)

				return perms.NextToken, nil
			})
			if err != nil {
				return nil, err
			}
			if len(principals) > 0 {
				r.AddAttribute("Permissions", principals)
			}

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
