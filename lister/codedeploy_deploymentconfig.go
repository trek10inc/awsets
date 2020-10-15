package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/trek10inc/awsets/option"
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

func (l AWSCodeDeployDeploymentConfig) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := codedeploy.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListDeploymentConfigs(cfg.Context, &codedeploy.ListDeploymentConfigsInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, config := range res.DeploymentConfigsList {
			configRes, err := svc.GetDeploymentConfig(cfg.Context, &codedeploy.GetDeploymentConfigInput{
				DeploymentConfigName: config,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get codedeploy deployment config %s: %w", *config, err)
			}
			v := configRes.DeploymentConfigInfo
			if v == nil {
				continue
			}
			r := resource.New(cfg, resource.CodeDeployDeploymentConfig, v.DeploymentConfigId, v.DeploymentConfigName, v)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
