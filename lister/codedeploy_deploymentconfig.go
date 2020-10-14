package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSCodeDeployDeploymentConfig struct {
}

func init() {
	i := AWSCodeDeployDeploymentConfig{}
	listers = append(listers, i)
}

func (l AWSCodeDeployDeploymentConfig) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.CodeDeployDeploymentConfig,
	}
}

func (l AWSCodeDeployDeploymentConfig) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := codedeploy.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListDeploymentConfigs(ctx.Context, &codedeploy.ListDeploymentConfigsInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, config := range res.DeploymentConfigsList {
			configRes, err := svc.GetDeploymentConfig(ctx.Context, &codedeploy.GetDeploymentConfigInput{
				DeploymentConfigName: config,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get codedeploy deployment config %s: %w", *config, err)
			}
			v := configRes.DeploymentConfigInfo
			if v == nil {
				continue
			}
			r := resource.New(ctx, resource.CodeDeployDeploymentConfig, v.DeploymentConfigId, v.DeploymentConfigName, v)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
