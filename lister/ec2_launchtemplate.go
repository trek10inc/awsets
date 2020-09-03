package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
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
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeLaunchTemplatesRequest(&ec2.DescribeLaunchTemplatesInput{
		MaxResults: aws.Int64(200),
	})

	rg := resource.NewGroup()
	res, err := req.Send(ctx.Context)
	if err != nil {
		return rg, err
	}

	for _, v := range res.LaunchTemplates {
		launchTemplates, err := svc.DescribeLaunchTemplateVersionsRequest(&ec2.DescribeLaunchTemplateVersionsInput{
			LaunchTemplateId: v.LaunchTemplateId,
			Versions:         []string{fmt.Sprintf("%d", *v.LatestVersionNumber)},
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to describe launch template version for %s: %w", *v.LaunchTemplateName, err)
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
						r.AddRelation(resource.KmsKey, bm.Ebs.KmsKeyId, "")
						r.AddRelation(resource.Ec2Snapshot, bm.Ebs.SnapshotId, "")
					}
				}
			}
			rg.AddResource(r)
		}
	}
	return rg, nil
}
