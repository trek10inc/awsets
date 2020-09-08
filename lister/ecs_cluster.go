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
	return []resource.ResourceType{resource.EcsCluster, resource.EcsService, resource.EcsTask}
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

			listServicesRequest := svc.ListServicesRequest(&ecs.ListServicesInput{
				Cluster:    v.ClusterArn,
				MaxResults: aws.Int64(10),
			})
			servicesPaginator := ecs.NewListServicesPaginator(listServicesRequest)
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
					serviceResource.AddRelation(resource.EcsCluster, clusterArn.ResourceId, "")

					if arn.IsArnP(service.RoleArn) {
						roleArn := arn.ParseP(service.RoleArn)
						serviceResource.AddRelation(resource.IamRole, roleArn.ResourceId, roleArn.ResourceVersion)
					}
					if arn.IsArnP(service.TaskDefinition) {
						taskDefArn := arn.ParseP(service.TaskDefinition)
						serviceResource.AddRelation(resource.EcsTaskDefinition, taskDefArn.ResourceId, "")
					}
					if arn.IsArnP(service.RoleArn) { //TODO this seems to be just a name, not an ARN?
						roleArn := arn.ParseP(service.RoleArn)
						serviceResource.AddRelation(resource.IamRole, roleArn.ResourceId, roleArn.ResourceVersion)
						rg.AddResource(serviceResource)
					}
				}
			}
			listTasksRequest := svc.ListTasksRequest(&ecs.ListTasksInput{
				Cluster:    v.ClusterArn,
				MaxResults: aws.Int64(100),
			})
			tasksPaginator := ecs.NewListTasksPaginator(listTasksRequest)
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
					taskResource.AddRelation(resource.EcsCluster, clusterArn.ResourceId, "")
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
		}
	}
	err := paginator.Err()
	return rg, err
}
