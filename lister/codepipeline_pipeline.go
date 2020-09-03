package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/trek10inc/awsets/resource"
)

type AWSCodepipelinePipeline struct {
}

func init() {
	i := AWSCodepipelinePipeline{}
	listers = append(listers, i)
}

func (l AWSCodepipelinePipeline) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CodePipelinePipeline}
}

func (l AWSCodepipelinePipeline) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := codepipeline.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	req := svc.ListPipelinesRequest(&codepipeline.ListPipelinesInput{})

	paginator := codepipeline.NewListPipelinesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Pipelines {
			pipeline, err := svc.GetPipelineRequest(&codepipeline.GetPipelineInput{
				Name:    v.Name,
				Version: v.Version,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get pipeline %s: %w", *v.Name, err)
			}
			r := resource.New(ctx, resource.CodePipelinePipeline, v.Name, v.Name, v)
			r.AddAttribute("Metadata", pipeline.Metadata)
			r.AddAttribute("Pipeline", pipeline.Pipeline)
			rg.AddResource(r)
		}

	}
	err := paginator.Err()
	return rg, err
}
