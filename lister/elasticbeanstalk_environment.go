package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"
)

type AWSElasticBeanstalkEnvironment struct {
}

func init() {
	i := AWSElasticBeanstalkEnvironment{}
	listers = append(listers, i)
}

func (l AWSElasticBeanstalkEnvironment) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ElasticBeanstalkEnvironment}
}

func (l AWSElasticBeanstalkEnvironment) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := elasticbeanstalk.New(ctx.AWSCfg)

	rg := resource.NewGroup()

	var nextToken *string

	for {
		envs, err := svc.DescribeEnvironmentsRequest(&elasticbeanstalk.DescribeEnvironmentsInput{
			MaxRecords: aws.Int64(100),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list elastic beanstalk environments: %w", err)
		}
		for _, v := range envs.Environments {
			r := resource.New(ctx, resource.ElasticBeanstalkEnvironment, v.EnvironmentId, v.EnvironmentName, v)
			r.AddRelation(resource.ElasticBeanstalkApplication, v.ApplicationName, "")
			// TODO: relationship to load balancer?

			opts, err := svc.DescribeConfigurationOptionsRequest(&elasticbeanstalk.DescribeConfigurationOptionsInput{
				ApplicationName: v.ApplicationName,
				EnvironmentName: v.EnvironmentName,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get configuration options for environment %s: %w", *v.EnvironmentName, err)
			}
			r.AddAttribute("ConfigurationOptions", opts.Options)
			settings, err := svc.DescribeConfigurationSettingsRequest(&elasticbeanstalk.DescribeConfigurationSettingsInput{
				ApplicationName: v.ApplicationName,
				EnvironmentName: v.EnvironmentName,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get configuration settings for environment %s: %w", *v.EnvironmentName, err)
			}
			r.AddAttribute("ConfigurationSettings", settings.ConfigurationSettings)

			rg.AddResource(r)
		}

		if envs.NextToken == nil {
			break
		}
		nextToken = envs.NextToken
	}

	return rg, nil
}
