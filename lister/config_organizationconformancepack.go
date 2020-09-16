package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/service/configservice"

	"github.com/trek10inc/awsets/context"

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

func (l AWSConfigOrganizationConformancePack) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := configservice.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextToken *string
	for {
		packs, err := svc.DescribeOrganizationConformancePacksRequest(&configservice.DescribeOrganizationConformancePacksInput{
			NextToken: nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list config organization conformance packs: %w", err)
		}
		for _, v := range packs.OrganizationConformancePacks {
			packArn := arn.ParseP(v.OrganizationConformancePackArn)
			r := resource.New(ctx, resource.ConfigOrganizationConformancePack, packArn.ResourceId, v.OrganizationConformancePackName, v)
			r.AddRelation(resource.S3Bucket, v.DeliveryS3Bucket, "")
			rg.AddResource(r)
		}
		if packs.NextToken == nil {
			break
		}
		nextToken = packs.NextToken
	}
	return rg, nil
}
