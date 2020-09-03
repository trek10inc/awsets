package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSCodepipelineWebhook struct {
}

func init() {
	i := AWSCodepipelineWebhook{}
	listers = append(listers, i)
}

func (l AWSCodepipelineWebhook) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CodePipelineWebhook}
}

func (l AWSCodepipelineWebhook) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := codepipeline.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	req := svc.ListWebhooksRequest(&codepipeline.ListWebhooksInput{})

	paginator := codepipeline.NewListWebhooksPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Webhooks {
			whArn := arn.ParseP(v.Arn)
			r := resource.New(ctx, resource.CodePipelineWebhook, whArn.ResourceId, whArn.ResourceId, v)
			if v.Definition != nil {
				r.AddRelation(resource.CodePipelinePipeline, v.Definition.TargetPipeline, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
