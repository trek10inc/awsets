package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSElasticBeanstalkApplication struct {
}

func init() {
	i := AWSElasticBeanstalkApplication{}
	listers = append(listers, i)
}

func (l AWSElasticBeanstalkApplication) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ElasticBeanstalkApplication}
}

func (l AWSElasticBeanstalkApplication) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := elasticbeanstalk.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()

	apps, err := svc.DescribeApplications(cfg.Context, &elasticbeanstalk.DescribeApplicationsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list elastic beanstalk applications: %w", err)
	}
	for _, v := range apps.Applications {
		r := resource.New(cfg, resource.ElasticBeanstalkApplication, v.ApplicationName, v.ApplicationName, v)
		rg.AddResource(r)
	}

	return rg, nil
}
