package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSServiceDiscoveryService struct {
}

func init() {
	i := AWSServiceDiscoveryService{}
	listers = append(listers, i)
}

func (l AWSServiceDiscoveryService) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ServiceDiscoveryService,
	}
}

func (l AWSServiceDiscoveryService) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := servicediscovery.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListServices(ctx.Context, &servicediscovery.ListServicesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, service := range res.Services {
			v, err := svc.GetService(ctx.Context, &servicediscovery.GetServiceInput{
				Id: service.Id,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe service discovery service %s: %w", *service.Id, err)
			}
			r := resource.New(ctx, resource.ServiceDiscoveryService, v.Service.Id, v.Service.Name, v.Service)
			r.AddRelation(resource.ServiceDiscoveryNamespace, v.Service.NamespaceId, "")

			// Service Discovery Instances
			err = Paginator(func(nt2 *string) (*string, error) {
				instances, err := svc.ListInstances(ctx.Context, &servicediscovery.ListInstancesInput{
					ServiceId:  service.Id,
					MaxResults: aws.Int32(100),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get instances for service %s: %w", *service.Id, err)
				}
				for _, instance := range instances.Instances {
					instanceR := resource.New(ctx, resource.ServiceDiscoveryInstance, instance.Id, instance.Id, instance)
					instanceR.AddRelation(resource.ServiceDiscoveryService, service.Id, "")

					if val, ok := instance.Attributes["AWS_EC2_INSTANCE_ID"]; ok {
						instanceR.AddRelation(resource.Ec2Instance, val, "")
					}

					rg.AddResource(instanceR)
				}

				return instances.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
