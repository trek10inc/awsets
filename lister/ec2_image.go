package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2Image struct {
}

func init() {
	i := AWSEc2Image{}
	listers = append(listers, i)
}

func (l AWSEc2Image) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2Image}
}

func (l AWSEc2Image) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := ec2.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()

	images, err := svc.DescribeImages(cfg.Context, &ec2.DescribeImagesInput{
		Owners: []*string{&cfg.AccountId},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list ec2 images: %w", err)
	}
	for _, image := range images.Images {
		r := resource.New(cfg, resource.Ec2Image, image.ImageId, image.Name, image)
		for _, bm := range image.BlockDeviceMappings {
			if bm.Ebs != nil {
				r.AddRelation(resource.KmsKey, bm.Ebs.KmsKeyId, "")
				r.AddRelation(resource.Ec2Snapshot, bm.Ebs.SnapshotId, "")
			}
		}
		rg.AddResource(r)
	}

	return rg, nil
}
