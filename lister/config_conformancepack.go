package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/trek10inc/awsets/option"
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

func (l AWSConfigConformancePack) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := configservice.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeConformancePacks(cfg.Context, &configservice.DescribeConformancePacksInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list config conformance packs: %w", err)
		}
		for _, v := range res.ConformancePackDetails {
			r := resource.New(cfg, resource.ConfigConformancePack, v.ConformancePackId, v.ConformancePackId, v)
			r.AddRelation(resource.S3Bucket, v.DeliveryS3Bucket, "")
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
