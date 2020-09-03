package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"

	"github.com/trek10inc/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSBackupVault struct {
}

func init() {
	i := AWSBackupVault{}
	listers = append(listers, i)
}

func (l AWSBackupVault) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.BackupVault,
	}
}

func (l AWSBackupVault) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := backup.New(ctx.AWSCfg)

	req := svc.ListBackupVaultsRequest(&backup.ListBackupVaultsInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := backup.NewListBackupVaultsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.BackupVaultList {

			vaultArn := arn.ParseP(v.BackupVaultArn)
			r := resource.New(ctx, resource.BackupVault, vaultArn.ResourceId, v.BackupVaultName, v)

			accessPolicy, err := svc.GetBackupVaultAccessPolicyRequest(&backup.GetBackupVaultAccessPolicyInput{
				BackupVaultName: v.BackupVaultName,
			}).Send(ctx.Context)
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					if aerr.Code() == backup.ErrCodeResourceNotFoundException &&
						strings.Contains(aerr.Message(), "has no associated policy") {
						// vaults may not have policy
						continue
					}
				}
				return rg, fmt.Errorf("failed to get access policy for vault %s: %w", *v.BackupVaultName, err)
			}
			r.AddAttribute("AccessPolicy", accessPolicy.GetBackupVaultAccessPolicyOutput)

			notifications, err := svc.GetBackupVaultNotificationsRequest(&backup.GetBackupVaultNotificationsInput{
				BackupVaultName: v.BackupVaultName,
			}).Send(ctx.Context)
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					if aerr.Code() == backup.ErrCodeResourceNotFoundException &&
						strings.Contains(aerr.Message(), "Failed reading notifications from database for Backup vault") {
						// vaults may not have notifications
						continue
					}
				}
				return rg, fmt.Errorf("failed to get notifications for vault %s: %w", *v.BackupVaultName, err)
			}

			r.AddAttribute("Notifications", notifications.GetBackupVaultNotificationsOutput)
			r.AddARNRelation(resource.SnsTopic, notifications.SNSTopicArn)

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
