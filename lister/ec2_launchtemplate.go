package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2LaunchTemplate struct {
}

func init() {
	i := AWSEc2LaunchTemplate{}
	listers = append(listers, i)
}

func (l AWSEc2LaunchTemplate) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2LaunchTemplate}
}

func (l AWSEc2LaunchTemplate) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeLaunchTemplates(ctx.Context, &ec2.DescribeLaunchTemplatesInput{
			MaxResults: aws.Int32(200),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.LaunchTemplates {
			launchTemplates, err := svc.DescribeLaunchTemplateVersions(ctx.Context, &ec2.DescribeLaunchTemplateVersionsInput{
				LaunchTemplateId: v.LaunchTemplateId,
				Versions:         []*string{aws.String(fmt.Sprintf("%d", *v.LatestVersionNumber))},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe launch template version for %s: %w", *v.LaunchTemplateName, err)
			}
			for _, lt := range launchTemplates.LaunchTemplateVersions {
				r := resource.New(ctx, resource.Ec2LaunchTemplate, lt.LaunchTemplateId, lt.LaunchTemplateName, lt)
				if data := lt.LaunchTemplateData; data != nil {
					//r.AddRelation(resource.Ec2Im)data.ImageId
					for _, ni := range data.NetworkInterfaces {
						r.AddRelation(resource.Ec2Subnet, ni.SubnetId, "")
						r.AddRelation(resource.Ec2NetworkInterface, ni.NetworkInterfaceId, "")
					}
					for _, sg := range data.SecurityGroupIds {
						r.AddRelation(resource.Ec2SecurityGroup, sg, "")
					}
					r.AddRelation(resource.Ec2KeyPair, data.KeyName, "")
					for _, bm := range data.BlockDeviceMappings {
						if bm.Ebs != nil {
							r.AddARNRelation(resource.KmsKey, bm.Ebs.KmsKeyId)
							r.AddRelation(resource.Ec2Snapshot, bm.Ebs.SnapshotId, "")
						}
					}
				}
				rg.AddResource(r)
			}
		}
		return res.NextToken, nil
	})

	return rg, err
}
