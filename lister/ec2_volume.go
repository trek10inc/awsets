package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
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

func (l AWSEc2Volume) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.New(ctx.AWSCfg)

	req := svc.DescribeVolumesRequest(&ec2.DescribeVolumesInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := ec2.NewDescribeVolumesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Volumes {
			r := resource.New(ctx, resource.Ec2Volume, v.VolumeId, v.VolumeId, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			for _, va := range v.Attachments {
				r.AddRelation(resource.Ec2Instance, va.InstanceId, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
