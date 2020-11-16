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

	svc := elasticbeanstalk.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeEnvironments(ctx.Context, &elasticbeanstalk.DescribeEnvironmentsInput{
			MaxRecords: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list elastic beanstalk environments: %w", err)
		}
		for _, v := range res.Environments {
			r := resource.New(ctx, resource.ElasticBeanstalkEnvironment, v.EnvironmentId, v.EnvironmentName, v)
			r.AddRelation(resource.ElasticBeanstalkApplication, v.ApplicationName, "")
			// TODO: relationship to load balancer?

			// Configuration Options
			opts, err := svc.DescribeConfigurationOptions(ctx.Context, &elasticbeanstalk.DescribeConfigurationOptionsInput{
				ApplicationName: v.ApplicationName,
				EnvironmentName: v.EnvironmentName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get configuration options for environment %s: %w", *v.EnvironmentName, err)
			}
			r.AddAttribute("ConfigurationOptions", opts.Options)

			// Configuration Settings
			settings, err := svc.DescribeConfigurationSettings(ctx.Context, &elasticbeanstalk.DescribeConfigurationSettingsInput{
				ApplicationName: v.ApplicationName,
				EnvironmentName: v.EnvironmentName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get configuration settings for environment %s: %w", *v.EnvironmentName, err)
			}
			r.AddAttribute("ConfigurationSettings", settings.ConfigurationSettings)

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
