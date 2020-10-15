package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/trek10inc/awsets/option"
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

func (l AWSCodepipelinePipeline) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := codepipeline.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListPipelines(cfg.Context, &codepipeline.ListPipelinesInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.Pipelines {
			pipeline, err := svc.GetPipeline(cfg.Context, &codepipeline.GetPipelineInput{
				Name:    v.Name,
				Version: v.Version,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get pipeline %s: %w", *v.Name, err)
			}
			r := resource.New(cfg, resource.CodePipelinePipeline, v.Name, v.Name, v)
			r.AddAttribute("Metadata", pipeline.Metadata)
			r.AddAttribute("Pipeline", pipeline.Pipeline)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
