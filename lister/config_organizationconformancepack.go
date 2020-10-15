package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSConfigOrganizationConformancePack struct {
}

func init() {
	i := AWSConfigOrganizationConformancePack{}
	listers = append(listers, i)
}

func (l AWSConfigOrganizationConformancePack) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ConfigOrganizationConformancePack}
}

func (l AWSConfigOrganizationConformancePack) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := configservice.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeOrganizationConformancePacks(cfg.Context, &configservice.DescribeOrganizationConformancePacksInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list config organization conformance packs: %w", err)
		}
		for _, v := range res.OrganizationConformancePacks {
			packArn := arn.ParseP(v.OrganizationConformancePackArn)
			r := resource.New(cfg, resource.ConfigOrganizationConformancePack, packArn.ResourceId, v.OrganizationConformancePackName, v)
			r.AddRelation(resource.S3Bucket, v.DeliveryS3Bucket, "")
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
