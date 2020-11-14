package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
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

	svc := sfn.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListStateMachines(ctx.Context, &sfn.ListStateMachinesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, sm := range res.StateMachines {

			res, err := svc.DescribeStateMachine(ctx.Context, &sfn.DescribeStateMachineInput{
				StateMachineArn: sm.StateMachineArn,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get state machine %s: %w", *sm.Name, err)
			}
			smArn := arn.ParseP(res.StateMachineArn)
			r := resource.New(ctx, resource.StepFunctionStateMachine, smArn.ResourceId, sm.Name, res)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
