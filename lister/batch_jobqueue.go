package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/batch"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	svc := batch.New(ctx.AWSCfg)

	req := svc.DescribeJobQueuesRequest(&batch.DescribeJobQueuesInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := batch.NewDescribeJobQueuesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.JobQueues {
			r := resource.New(ctx, resource.BatchJobQueue, v.JobQueueName, v.JobQueueName, v)
			for _, ce := range v.ComputeEnvironmentOrder {
				r.AddARNRelation(resource.BatchComputeEnvironment, ce.ComputeEnvironment)
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
