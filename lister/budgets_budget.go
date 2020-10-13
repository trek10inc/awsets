package lister

import (
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
	svc := budgets.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeBudgets(ctx.Context, &budgets.DescribeBudgetsInput{
			AccountId:  &ctx.AccountId,
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, budget := range res.Budgets {
			r := resource.New(ctx, resource.BudgetsBudget, budget.BudgetName, budget.BudgetName, budget)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
