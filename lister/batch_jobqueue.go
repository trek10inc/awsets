package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSBatchJobQueue struct {
}

func init() {
	i := AWSBatchJobQueue{}
	listers = append(listers, i)
}

func (l AWSBatchJobQueue) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.BatchJobQueue}
}

func (l AWSBatchJobQueue) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := batch.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeJobQueues(ctx.Context, &batch.DescribeJobQueuesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.JobQueues {
			r := resource.New(ctx, resource.BatchJobQueue, v.JobQueueName, v.JobQueueName, v)
			for _, ce := range v.ComputeEnvironmentOrder {
				r.AddARNRelation(resource.BatchComputeEnvironment, ce.ComputeEnvironment)
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
