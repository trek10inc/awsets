package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/batch"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSBatchJobDefinition struct {
}

func init() {
	i := AWSBatchJobDefinition{}
	listers = append(listers, i)
}

func (l AWSBatchJobDefinition) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.BatchJobDefinition}
}

func (l AWSBatchJobDefinition) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := batch.New(ctx.AWSCfg)

	req := svc.DescribeJobDefinitionsRequest(&batch.DescribeJobDefinitionsInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := batch.NewDescribeJobDefinitionsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.JobDefinitions {
			r := resource.New(ctx, resource.BatchJobDefinition, v.JobDefinitionName, v.JobDefinitionName, v)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
