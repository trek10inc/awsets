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

	svc := codepipeline.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListWebhooks(ctx.Context, &codepipeline.ListWebhooksInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.Webhooks {
			whArn := arn.ParseP(v.Arn)
			r := resource.New(ctx, resource.CodePipelineWebhook, whArn.ResourceId, whArn.ResourceId, v)
			if v.Definition != nil {
				r.AddRelation(resource.CodePipelinePipeline, v.Definition.TargetPipeline, "")
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
