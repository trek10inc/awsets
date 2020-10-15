package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/batch"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/option"
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

func (l AWSBatchJobDefinition) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := batch.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeJobDefinitions(cfg.Context, &batch.DescribeJobDefinitionsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.JobDefinitions {
			r := resource.New(cfg, resource.BatchJobDefinition, v.JobDefinitionName, v.JobDefinitionName, v)
			rg.AddResource(r)
		}

		return res.NextToken, nil
	})
	return rg, err
}
