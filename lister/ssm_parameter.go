package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
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

func (l AWSSsmParameter) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ssm.NewFromConfig(ctx.AWSCfg)
	res, err := svc.DescribeParameters(ctx.Context, &ssm.DescribeParametersInput{
		MaxResults: aws.Int32(50),
	})

	rg := resource.NewGroup()
	paginator := ssm.NewDescribeParametersPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, parameter := range page.Parameters {
			r := resource.New(ctx, resource.SsmParameter, parameter.Name, parameter.Name, parameter)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
