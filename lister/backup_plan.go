package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSBackupPlan struct {
}

func init() {
	i := AWSBackupPlan{}
	listers = append(listers, i)
}

func (l AWSBackupPlan) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.BackupPlan,
		resource.BackupSelection,
	}
}

func (l AWSBackupPlan) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := backup.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListBackupPlans(ctx.Context, &backup.ListBackupPlansInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, plan := range res.BackupPlansList {
			v, err := svc.GetBackupPlan(ctx.Context, &backup.GetBackupPlanInput{
				BackupPlanId: plan.BackupPlanId,
				VersionId:    plan.VersionId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get backup plan %s: %w", *plan.BackupPlanId, err)
			}
			r := resource.New(ctx, resource.BackupPlan, v.BackupPlanId, v.BackupPlanId, v)

			err = Paginator(func(nt2 *string) (*string, error) {
				selectionsRes, err := svc.ListBackupSelections(ctx.Context, &backup.ListBackupSelectionsInput{
					BackupPlanId: plan.BackupPlanId,
					MaxResults:   aws.Int32(25),
					NextToken:    nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list selections for plan %s: %w", *v.BackupPlanId, err)
				}
				for _, selectionId := range selectionsRes.BackupSelectionsList {
					selection, err := svc.GetBackupSelection(ctx.Context, &backup.GetBackupSelectionInput{
						BackupPlanId: v.BackupPlanId,
						SelectionId:  selectionId.SelectionId,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to get selection %s for plan %s: %w", *selectionId.SelectionId, *plan.BackupPlanId, err)
					}
					selectionR := resource.New(ctx, resource.BackupSelection, selection.SelectionId, selection.SelectionId, selection)
					selectionR.AddRelation(resource.BackupPlan, v.BackupPlanId, v.VersionId)

					rg.AddResource(selectionR)
				}

				return selectionsRes.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
