package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/trek10inc/awsets/arn"
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
	svc := backup.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListBackupVaults(ctx.Context, &backup.ListBackupVaultsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.BackupVaultList {

			vaultArn := arn.ParseP(v.BackupVaultArn)
			r := resource.New(ctx, resource.BackupVault, vaultArn.ResourceId, v.BackupVaultName, v)

			accessPolicy, err := svc.GetBackupVaultAccessPolicy(ctx.Context, &backup.GetBackupVaultAccessPolicyInput{
				BackupVaultName: v.BackupVaultName,
			})
			if err != nil {
				if strings.Contains(err.Error(), "has no associated policy") {
					// vaults may not have policy
					continue
				}
				return nil, fmt.Errorf("failed to get access policy for vault %s: %w", *v.BackupVaultName, err)
			}
			r.AddAttribute("AccessPolicy", accessPolicy)

			notifications, err := svc.GetBackupVaultNotifications(ctx.Context, &backup.GetBackupVaultNotificationsInput{
				BackupVaultName: v.BackupVaultName,
			})
			if err != nil {
				if strings.Contains(err.Error(), "Failed reading notifications from database for Backup vault") {
					// vaults may not have notifications
					continue
				}
				return nil, fmt.Errorf("failed to get notifications for vault %s: %w", *v.BackupVaultName, err)
			}

			r.AddAttribute("Notifications", notifications)
			r.AddARNRelation(resource.SnsTopic, notifications.SNSTopicArn)

			rg.AddResource(r)
		}

		return res.NextToken, nil
	})
	return rg, err
}
