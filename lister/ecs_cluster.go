package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/trek10inc/awsets/arn"
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

func (l AWSEcsCluster) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ecs.New(ctx.AWSCfg)
	req := svc.ListClustersRequest(&ecs.ListClustersInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := ecs.NewListClustersPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		if len(page.ClusterArns) == 0 {
			continue
		}
		clustersRequest := svc.DescribeClustersRequest(&ecs.DescribeClustersInput{
			Clusters: page.ClusterArns,
			Include:  []ecs.ClusterField{ecs.ClusterFieldAttachments, ecs.ClusterFieldSettings, ecs.ClusterFieldStatistics, ecs.ClusterFieldTags},
		})
		res, err := clustersRequest.Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to send cluster request: %w", err)
		}
		for _, v := range res.Clusters {
			clusterArn := arn.ParseP(v.ClusterArn)
			clusterResource := resource.New(ctx, resource.EcsCluster, clusterArn.ResourceId, v.ClusterName, v)

			// ECS Services
			servicesPaginator := ecs.NewListServicesPaginator(svc.ListServicesRequest(&ecs.ListServicesInput{
				Cluster:    v.ClusterArn,
				MaxResults: aws.Int64(10),
			}))
			for servicesPaginator.Next(ctx.Context) {
				servicesPage := servicesPaginator.CurrentPage()
				if len(servicesPage.ServiceArns) == 0 {
					continue
				}
				describeServicesReq := svc.DescribeServicesRequest(&ecs.DescribeServicesInput{
					Cluster:  v.ClusterArn,
					Include:  []ecs.ServiceField{ecs.ServiceFieldTags},
					Services: servicesPage.ServiceArns,
				})
				servicesResponse, err := describeServicesReq.Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to send describe services request: %w", err)
				}
				for _, service := range servicesResponse.Services {

					serviceArn := arn.ParseP(service.ServiceArn)
					serviceResource := resource.New(ctx, resource.EcsService, serviceArn.ResourceId, service.ServiceName, service)
					serviceResource.AddARNRelation(resource.EcsCluster, v.ClusterArn)
					serviceResource.AddARNRelation(resource.IamRole, service.RoleArn)
					serviceResource.AddARNRelation(resource.EcsTaskDefinition, service.TaskDefinition)

					if arn.IsArnP(service.RoleArn) { //TODO this seems to be just a name, not an ARN?
						roleArn := arn.ParseP(service.RoleArn)
						serviceResource.AddRelation(resource.IamRole, roleArn.ResourceId, roleArn.ResourceVersion)
						rg.AddResource(serviceResource)
					}
				}
			}

			// ECS Tasks
			tasksPaginator := ecs.NewListTasksPaginator(svc.ListTasksRequest(&ecs.ListTasksInput{
				Cluster:    v.ClusterArn,
				MaxResults: aws.Int64(100),
			}))
			for tasksPaginator.Next(ctx.Context) {
				tasksPage := tasksPaginator.CurrentPage()
				if len(tasksPage.TaskArns) == 0 {
					continue
				}
				describeTasksReq := svc.DescribeTasksRequest(&ecs.DescribeTasksInput{
					Cluster: v.ClusterArn,
					Include: []ecs.TaskField{ecs.TaskFieldTags},
					Tasks:   tasksPage.TaskArns,
				})
				tasksResponse, err := describeTasksReq.Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to send describe services request: %w", err)
				}
				for _, task := range tasksResponse.Tasks {

					taskArn := arn.ParseP(task.TaskArn)
					taskResource := resource.New(ctx, resource.EcsTask, taskArn.ResourceId, "", task)
					taskResource.AddARNRelation(resource.EcsCluster, v.ClusterArn)
					taskDefArn := arn.ParseP(task.TaskDefinitionArn)
					taskResource.AddRelation(resource.EcsTaskDefinition, taskDefArn.ResourceId, "")
					rg.AddResource(taskResource)
				}
			}
			rg.AddResource(clusterResource)
			err = servicesPaginator.Err()
			if err != nil {
				return rg, fmt.Errorf("failed to paginate services: %w", err)
			}

			// ECS Capacity Providers
			var nextToken *string
			for {
				providers, err := svc.DescribeCapacityProvidersRequest(&ecs.DescribeCapacityProvidersInput{
					CapacityProviders: v.CapacityProviders,
					Include:           []ecs.CapacityProviderField{ecs.CapacityProviderFieldTags},
					NextToken:         nextToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to describe ecs capacity providers for %s: %w", *v.ClusterName, err)
				}
				for _, provider := range providers.CapacityProviders {
					providerArn := arn.ParseP(provider.CapacityProviderArn)
					providerRes := resource.New(ctx, resource.EcsCapacityProvider, providerArn.ResourceId, provider.Name, provider)
					providerRes.AddARNRelation(resource.EcsCluster, v.ClusterArn)
					rg.AddResource(providerRes)
				}
				if providers.NextToken == nil {
					break
				}
				nextToken = providers.NextToken
			}
			// TODO: task sets?
		}
	}
	err := paginator.Err()
	return rg, err
}
