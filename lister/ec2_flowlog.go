package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AWSEc2FlowLog struct {
}

func init() {
	i := AWSEc2FlowLog{}
	listers = append(listers, i)
}

func (l AWSEc2FlowLog) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2FlowLog}
}

func (l AWSEc2FlowLog) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeFlowLogsRequest(&ec2.DescribeFlowLogsInput{
		MaxResults: aws.Int64(1000),
	})

	rg := resource.NewGroup()
	paginator := ec2.NewDescribeFlowLogsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.FlowLogs {
			r := resource.New(ctx, resource.Ec2FlowLog, v.FlowLogId, v.FlowLogId, v)
			r.AddRelation(resource.LogGroup, v.LogGroupName, "")
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
