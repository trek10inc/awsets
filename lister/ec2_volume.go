package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2Volume struct {
}

func init() {
	i := AWSEc2Volume{}
	listers = append(listers, i)
}

func (l AWSEc2Volume) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2Volume}
}

func (l AWSEc2Volume) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ec2.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeVolumes(cfg.Context, &ec2.DescribeVolumesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.Volumes {
			r := resource.New(cfg, resource.Ec2Volume, v.VolumeId, v.VolumeId, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			for _, va := range v.Attachments {
				r.AddRelation(resource.Ec2Instance, va.InstanceId, "")
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
