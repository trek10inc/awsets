package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/batch"

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
	svc := batch.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeJobDefinitions(ctx.Context, &batch.DescribeJobDefinitionsInput{
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.JobDefinitions {
			r := resource.New(ctx, resource.BatchJobDefinition, v.JobDefinitionName, v.JobDefinitionName, v)
			rg.AddResource(r)
		}

		return res.NextToken, nil
	})
	return rg, err
}
