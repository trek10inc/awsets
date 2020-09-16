package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/configservice"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"
)

type AWSConfigConformancePack struct {
}

func init() {
	i := AWSConfigConformancePack{}
	listers = append(listers, i)
}

func (l AWSConfigConformancePack) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ConfigConformancePack}
}

func (l AWSConfigConformancePack) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := configservice.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	var nextToken *string
	for {
		packs, err := svc.DescribeConformancePacksRequest(&configservice.DescribeConformancePacksInput{
			NextToken: nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list config conformance packs: %w", err)
		}
		for _, v := range packs.ConformancePackDetails {
			r := resource.New(ctx, resource.ConfigConformancePack, v.ConformancePackId, v.ConformancePackId, v)
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
