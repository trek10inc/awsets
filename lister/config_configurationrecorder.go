package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSConfigConfigurationRecorder struct {
}

func init() {
	i := AWSConfigConfigurationRecorder{}
	listers = append(listers, i)
}

func (l AWSConfigConfigurationRecorder) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ConfigConfigurationRecorder}
}

func (l AWSConfigConfigurationRecorder) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := configservice.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()

	recorders, err := svc.DescribeConfigurationRecorders(cfg.Context, &configservice.DescribeConfigurationRecordersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configuration recorders: %w", err)
	}
	for _, v := range recorders.ConfigurationRecorders {
		r := resource.New(cfg, resource.ConfigConfigurationRecorder, v.Name, v.Name, v)
		if v.RoleARN != nil {
			roleArn := arn.ParseP(v.RoleARN)
			r.AddRelation(resource.IamRole, roleArn.ResourceId, roleArn.ResourceVersion)
		}
		rg.AddResource(r)
	}
	return rg, nil
}
