package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloud9"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSCloud9Environment struct {
}

func init() {
	i := AWSCloud9Environment{}
	listers = append(listers, i)
}

func (l AWSCloud9Environment) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Cloud9Environment}
}

func (l AWSCloud9Environment) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := cloud9.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListEnvironments(cfg.Context, &cloud9.ListEnvironmentsInput{
			MaxResults: aws.Int32(25),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		if len(res.EnvironmentIds) == 0 {
			return nil, nil
		}
		environments, err := svc.DescribeEnvironments(cfg.Context, &cloud9.DescribeEnvironmentsInput{
			EnvironmentIds: res.EnvironmentIds,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe environments: %w", err)
		}
		for _, v := range environments.Environments {
			r := resource.New(cfg, resource.Cloud9Environment, v.Name, v.Name, v)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
