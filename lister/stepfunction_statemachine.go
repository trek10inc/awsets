package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/resource"
)

type AWSStepFunctionStateMachine struct {
}

func init() {
	i := AWSStepFunctionStateMachine{}
	listers = append(listers, i)
}

func (l AWSStepFunctionStateMachine) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.StepFunctionStateMachine}
}

func (l AWSStepFunctionStateMachine) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := sfn.New(ctx.AWSCfg)

	req := svc.ListStateMachinesRequest(&sfn.ListStateMachinesInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := sfn.NewListStateMachinesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, sm := range page.StateMachines {

			res, err := svc.DescribeStateMachineRequest(&sfn.DescribeStateMachineInput{
				StateMachineArn: sm.StateMachineArn,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get state machine %s: %w", *sm.Name, err)
			}
			smArn := arn.ParseP(res.StateMachineArn)
			r := resource.New(ctx, resource.StepFunctionStateMachine, smArn.ResourceId, sm.Name, res.DescribeStateMachineOutput)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
