package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSEcsCluster struct {
}

func init() {
	i := AWSEcsCluster{}
	listers = append(listers, i)
}

func (l AWSEcsCluster) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.EcsCluster,
		resource.EcsService,
		resource.EcsTask,
		resource.EcsCapacityProvider,
	}
}

func (l AWSEcsCluster) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ecs.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListClusters(cfg.Context, &ecs.ListClustersInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		if len(res.ClusterArns) == 0 {
			return nil, nil
		}
		clusters, err := svc.DescribeClusters(cfg.Context, &ecs.DescribeClustersInput{
			Clusters: res.ClusterArns,
			Include: []types.ClusterField{
				types.ClusterFieldAttachments,
				types.ClusterFieldSettings,
				types.ClusterFieldStatistics,
				types.ClusterFieldTags,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send cluster request: %w", err)
		}
		for _, v := range clusters.Clusters {
			clusterArn := arn.ParseP(v.ClusterArn)
			clusterResource := resource.New(cfg, resource.EcsCluster, clusterArn.ResourceId, v.ClusterName, v)
			rg.AddResource(clusterResource)

			// ECS Services
			err = Paginator(func(nt2 *string) (*string, error) {
				services, err := svc.ListServices(cfg.Context, &ecs.ListServicesInput{
					Cluster:    v.ClusterArn,
					MaxResults: aws.Int32(10),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to send list services request: %w", err)
				}
				if len(services.ServiceArns) == 0 {
					return nil, nil
				}
				describeServices, err := svc.DescribeServices(cfg.Context, &ecs.DescribeServicesInput{
					Cluster:  v.ClusterArn,
					Include:  []types.ServiceField{types.ServiceFieldTags},
					Services: services.ServiceArns,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to send describe services request: %w", err)
				}
				for _, service := range describeServices.Services {

					serviceArn := arn.ParseP(service.ServiceArn)
					serviceResource := resource.New(cfg, resource.EcsService, serviceArn.ResourceId, service.ServiceName, service)
					serviceResource.AddARNRelation(resource.EcsCluster, v.ClusterArn)
					serviceResource.AddARNRelation(resource.IamRole, service.RoleArn)
					serviceResource.AddARNRelation(resource.EcsTaskDefinition, service.TaskDefinition)

					if arn.IsArnP(service.RoleArn) { //TODO this seems to be just a name, not an ARN?
						roleArn := arn.ParseP(service.RoleArn)
						serviceResource.AddRelation(resource.IamRole, roleArn.ResourceId, roleArn.ResourceVersion)
						rg.AddResource(serviceResource)
					}
				}
				return services.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// ECS Tasks
			err = Paginator(func(nt2 *string) (*string, error) {
				tasks, err := svc.ListTasks(cfg.Context, &ecs.ListTasksInput{
					Cluster:    v.ClusterArn,
					MaxResults: aws.Int32(100),
				})
				if err != nil {
					return nil, fmt.Errorf("failed to send list services request: %w", err)
				}
				if len(tasks.TaskArns) == 0 {
					return nil, nil
				}
				describeTasks, err := svc.DescribeTasks(cfg.Context, &ecs.DescribeTasksInput{
					Cluster: v.ClusterArn,
					Include: []types.TaskField{types.TaskFieldTags},
					Tasks:   tasks.TaskArns,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to send describe tasks request: %w", err)
				}
				for _, task := range describeTasks.Tasks {
					taskArn := arn.ParseP(task.TaskArn)
					taskResource := resource.New(cfg, resource.EcsTask, taskArn.ResourceId, "", task)
					taskResource.AddARNRelation(resource.EcsCluster, v.ClusterArn)
					taskDefArn := arn.ParseP(task.TaskDefinitionArn)
					taskResource.AddRelation(resource.EcsTaskDefinition, taskDefArn.ResourceId, "")
					rg.AddResource(taskResource)
				}
				return tasks.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// ECS Capacity Providers
			err = Paginator(func(nt2 *string) (*string, error) {
				providers, err := svc.DescribeCapacityProviders(cfg.Context, &ecs.DescribeCapacityProvidersInput{
					CapacityProviders: v.CapacityProviders,
					Include:           []types.CapacityProviderField{types.CapacityProviderFieldTags},
					NextToken:         nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to describe ecs capacity providers for %s: %w", *v.ClusterName, err)
				}
				for _, provider := range providers.CapacityProviders {
					providerArn := arn.ParseP(provider.CapacityProviderArn)
					providerRes := resource.New(cfg, resource.EcsCapacityProvider, providerArn.ResourceId, provider.Name, provider)
					providerRes.AddARNRelation(resource.EcsCluster, v.ClusterArn)
					rg.AddResource(providerRes)
				}
				return res.NextToken, nil
			})
			if err != nil {
				return nil, err
			}
			// TODO: task sets?
		}
		return res.NextToken, nil
	})
	return rg, err
}
