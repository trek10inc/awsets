package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := ec2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeFlowLogs(ctx.Context, &ec2.DescribeFlowLogsInput{
			MaxResults: 1000,
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.FlowLogs {
			r := resource.New(ctx, resource.Ec2FlowLog, v.FlowLogId, v.FlowLogId, v)
			r.AddRelation(resource.LogGroup, v.LogGroupName, "")
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
