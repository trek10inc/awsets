package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/trek10inc/awsets/arn"
)

type AWSEcsTaskDefinition struct {
}

func init() {
	i := AWSEcsTaskDefinition{}
	listers = append(listers, i)
}

func (l AWSEcsTaskDefinition) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.EcsTaskDefinition}
}

func (l AWSEcsTaskDefinition) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ecs.New(ctx.AWSCfg)
	req := svc.ListTaskDefinitionsRequest(&ecs.ListTaskDefinitionsInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	taskDefinitionsPaginator := ecs.NewListTaskDefinitionsPaginator(req)
	for taskDefinitionsPaginator.Next(ctx.Context) {
		page := taskDefinitionsPaginator.CurrentPage()
		for _, taskDefArn := range page.TaskDefinitionArns {
			describeTaskDefReq := svc.DescribeTaskDefinitionRequest(&ecs.DescribeTaskDefinitionInput{
				Include:        []ecs.TaskDefinitionField{ecs.TaskDefinitionFieldTags},
				TaskDefinition: &taskDefArn,
			})
			describeTaskDefRes, err := describeTaskDefReq.Send(ctx.Context)
			if err != nil {
				return rg, err
			}
			parsedArn := arn.Parse(taskDefArn)
			taskDefResource := resource.New(ctx, resource.EcsTaskDefinition, parsedArn.ResourceId, "", describeTaskDefRes)
			rg.AddResource(taskDefResource)
		}
	}
	err := taskDefinitionsPaginator.Err()
	return rg, err
}
