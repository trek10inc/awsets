package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/glue"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSGlueWorkflow struct {
}

func init() {
	i := AWSGlueWorkflow{}
	listers = append(listers, i)
}

func (l AWSGlueWorkflow) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GlueWorkflow,
	}
}

func (l AWSGlueWorkflow) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := glue.NewFromConfig(ctx.AWSCfg)
	res, err := svc.ListWorkflows(ctx.Context, &glue.ListWorkflowsInput{
		MaxResults: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := glue.NewListWorkflowsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		wfRes, err := svc.BatchGetWorkflows(ctx.Context, &glue.BatchGetWorkflowsInput{
			IncludeGraph: aws.Bool(true),
			Names:        page.Workflows,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get glue workflows: %w", err)
		}
		for _, wf := range wfRes.Workflows {
			r := resource.New(ctx, resource.GlueWorkflow, wf.Name, wf.Name, wf)
			rg.AddResource(r)
			// TODO: explore nodes/edges
		}
	}

	err := paginator.Err()
	return rg, err
}
