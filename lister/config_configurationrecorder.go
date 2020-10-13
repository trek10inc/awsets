package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
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

func (l AWSConfigConfigurationRecorder) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := configservice.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	recorders, err := svc.DescribeConfigurationRecorders(ctx.Context, &configservice.DescribeConfigurationRecordersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configuration recorders: %w", err)
	}
	for _, v := range recorders.ConfigurationRecorders {
		r := resource.New(ctx, resource.ConfigConfigurationRecorder, v.Name, v.Name, v)
		if v.RoleARN != nil {
			roleArn := arn.ParseP(v.RoleARN)
			r.AddRelation(resource.IamRole, roleArn.ResourceId, roleArn.ResourceVersion)
		}
		rg.AddResource(r)
	}
	return rg, nil
}
