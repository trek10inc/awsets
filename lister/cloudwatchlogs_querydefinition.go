package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSCloudwatchLogsQueryDefinition struct {
}

func init() {
	i := AWSCloudwatchLogsQueryDefinition{}
	listers = append(listers, i)
}

func (l AWSCloudwatchLogsQueryDefinition) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.LogQueryDefinition,
	}
}

func (l AWSCloudwatchLogsQueryDefinition) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloudwatchlogs.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeQueryDefinitions(ctx.Context, &cloudwatchlogs.DescribeQueryDefinitionsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.QueryDefinitions {
			r := resource.New(ctx, resource.LogQueryDefinition, v.QueryDefinitionId, v.Name, v)
			for _, lg := range v.LogGroupNames {
				r.AddRelation(resource.LogGroup, lg, "")
			}
			rg.AddResource(r)
		}

		return res.NextToken, nil
	})
	return rg, err
}
