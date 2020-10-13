package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSGlueJob struct {
}

func init() {
	i := AWSGlueJob{}
	listers = append(listers, i)
}

func (l AWSGlueJob) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GlueJob,
		resource.GlueTrigger,
	}
}

func (l AWSGlueJob) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := glue.NewFromConfig(ctx.AWSCfg)
	res, err := svc.GetJobs(ctx.Context, &glue.GetJobsInput{
		MaxResults: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := glue.NewGetJobsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Jobs {

			r := resource.New(ctx, resource.GlueJob, v.Name, v.Name, v)
			r.AddARNRelation(resource.IamRole, v.Role)
			if v.Connections != nil {
				for _, c := range v.Connections.Connections {
					r.AddRelation(resource.GlueConnection, c, "")
				}
			}

			triggerPag := glue.NewGetTriggersPaginator(svc.GetTriggers(ctx.Context, &glue.GetTriggersInput{
				DependentJobName: v.Name,
				MaxResults:       aws.Int32(50),
			}))
			for triggerPag.Next(ctx.Context) {
				triggerPage := triggerPag.CurrentPage()
				for _, t := range triggerPage.Triggers {
					tRes := resource.New(ctx, resource.GlueTrigger, t.Id, t.Name, t)
					tRes.AddRelation(resource.GlueJob, v.Name, "")
					tRes.AddRelation(resource.GlueWorkflow, t.WorkflowName, "")
					rg.AddResource(tRes)
				}
			}
			err := triggerPag.Err()
			if err != nil {
				return nil, fmt.Errorf("failed to get triggers for job %s: %w", *v.Name, err)
			}
			rg.AddResource(r)
		}
	}

	err := paginator.Err()
	return rg, err
}
