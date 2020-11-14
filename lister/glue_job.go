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

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.GetJobs(ctx.Context, &glue.GetJobsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.Jobs {

			r := resource.New(ctx, resource.GlueJob, v.Name, v.Name, v)
			r.AddARNRelation(resource.IamRole, v.Role)
			if v.Connections != nil {
				for _, c := range v.Connections.Connections {
					r.AddRelation(resource.GlueConnection, c, "")
				}
			}

			// Triggers
			err = Paginator(func(nt2 *string) (*string, error) {
				triggers, err := svc.GetTriggers(ctx.Context, &glue.GetTriggersInput{
					DependentJobName: v.Name,
					MaxResults:       aws.Int32(50),
					NextToken:        nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get triggers for job %s: %w", *v.Name, err)
				}
				for _, t := range triggers.Triggers {
					tRes := resource.New(ctx, resource.GlueTrigger, t.Id, t.Name, t)
					tRes.AddRelation(resource.GlueJob, v.Name, "")
					tRes.AddRelation(resource.GlueWorkflow, t.WorkflowName, "")
					rg.AddResource(tRes)
				}
				return triggers.NextToken, nil
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
