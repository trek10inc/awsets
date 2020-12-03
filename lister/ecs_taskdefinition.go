package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := ecs.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListTaskDefinitions(ctx.Context, &ecs.ListTaskDefinitionsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		for _, taskDefArn := range res.TaskDefinitionArns {
			task, err := svc.DescribeTaskDefinition(ctx.Context, &ecs.DescribeTaskDefinitionInput{
				Include:        []types.TaskDefinitionField{types.TaskDefinitionFieldTags},
				TaskDefinition: &taskDefArn,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe task def %s: %w", taskDefArn, err)
			}
			parsedArn := arn.ParseP(task.TaskDefinition.TaskDefinitionArn)
			taskDefResource := resource.New(ctx, resource.EcsTaskDefinition, parsedArn.ResourceId, "", task.TaskDefinition)
			for _, v := range task.Tags {
				taskDefResource.Tags[aws.ToString(v.Key)] = aws.ToString(v.Value)
			}
			rg.AddResource(taskDefResource)
		}
		if err != nil {
			return nil, err
		}

		return res.NextToken, nil
	})
	return rg, err
}
