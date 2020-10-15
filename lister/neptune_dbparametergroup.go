package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSNeptuneDbParameterGroup struct {
}

func init() {
	i := AWSNeptuneDbParameterGroup{}
	listers = append(listers, i)
}

func (l AWSNeptuneDbParameterGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.NeptuneDbParameterGroup}
}

func (l AWSNeptuneDbParameterGroup) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := neptune.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeDBParameterGroups(cfg.Context, &neptune.DescribeDBParameterGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.DBParameterGroups {
			groupArn := arn.ParseP(v.DBParameterGroupArn)
			r := resource.New(cfg, resource.NeptuneDbParameterGroup, groupArn.ResourceId, "", v)
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
