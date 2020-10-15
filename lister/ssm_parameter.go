package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSSsmParameter struct {
}

func init() {
	i := AWSSsmParameter{}
	listers = append(listers, i)
}

func (l AWSSsmParameter) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.SsmParameter}
}

func (l AWSSsmParameter) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ssm.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeParameters(cfg.Context, &ssm.DescribeParametersInput{
			MaxResults: aws.Int32(50),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, parameter := range res.Parameters {
			r := resource.New(cfg, resource.SsmParameter, parameter.Name, parameter.Name, parameter)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
