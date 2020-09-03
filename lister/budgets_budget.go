package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/budgets"

	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSBudgetsBudget struct {
}

func init() {
	i := AWSBudgetsBudget{}
	listers = append(listers, i)
}

func (l AWSBudgetsBudget) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.BudgetsBudget}
}

func (l AWSBudgetsBudget) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := budgets.New(ctx.AWSCfg)

	rg := resource.NewGroup()

	var nextToken *string

	for {
		res, err := svc.DescribeBudgetsRequest(&budgets.DescribeBudgetsInput{
			AccountId:  &ctx.AccountId,
			MaxResults: aws.Int64(100),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list budgets: %w", err)
		}
		for _, budget := range res.Budgets {
			r := resource.New(ctx, resource.BudgetsBudget, budget.BudgetName, budget.BudgetName, budget)
			rg.AddResource(r)
		}
		if res.NextToken == nil {
			break
		}
		nextToken = res.NextToken
	}
	return rg, nil
}
