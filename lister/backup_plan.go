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
	svc := backup.New(ctx.AWSCfg)

	req := svc.ListBackupPlansRequest(&backup.ListBackupPlansInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := backup.NewListBackupPlansPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, plan := range page.BackupPlansList {
			v, err := svc.GetBackupPlanRequest(&backup.GetBackupPlanInput{
				BackupPlanId: plan.BackupPlanId,
				VersionId:    plan.VersionId,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get backup plan %s: %w", *plan.BackupPlanId, err)
			}
			r := resource.New(ctx, resource.BackupPlan, v.BackupPlanId, v.BackupPlanId, v)

			selectionPaginator := backup.NewListBackupSelectionsPaginator(svc.ListBackupSelectionsRequest(&backup.ListBackupSelectionsInput{
				BackupPlanId: plan.BackupPlanId,
				MaxResults:   aws.Int64(25),
			}))
			for selectionPaginator.Next(ctx.Context) {
				selectionPage := selectionPaginator.CurrentPage()
				for _, selectionId := range selectionPage.BackupSelectionsList {
					selection, err := svc.GetBackupSelectionRequest(&backup.GetBackupSelectionInput{
						BackupPlanId: v.BackupPlanId,
						SelectionId:  selectionId.SelectionId,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to get selection %s for plan %s: %w", *selectionId.SelectionId, *plan.BackupPlanId, err)
					}
					selectionR := resource.New(ctx, resource.BackupSelection, selection.SelectionId, selection.SelectionId, selection.GetBackupSelectionOutput)
					selectionR.AddRelation(resource.BackupPlan, v.BackupPlanId, v.VersionId)

					rg.AddResource(selectionR)
				}
			}
			if err = selectionPaginator.Err(); err != nil {
				return rg, fmt.Errorf("failed to list selections for plan %s: %w", *v.BackupPlanId, err)
			}

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
