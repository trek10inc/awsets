package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSSESTemplate struct {
}

func init() {
	i := AWSSESTemplate{}
	listers = append(listers, i)
}

func (l AWSSESTemplate) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.SesTemplate}
}

func (l AWSSESTemplate) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ses.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListTemplates(cfg.Context, &ses.ListTemplatesInput{
			MaxItems:  aws.Int32(10),
			NextToken: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, template := range res.TemplatesMetadata {
			v, err := svc.GetTemplate(cfg.Context, &ses.GetTemplateInput{
				TemplateName: template.Name,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get email template %s: %w", *template.Name, err)
			}
			r := resource.New(cfg, resource.SesTemplate, v.Template.TemplateName, v.Template.TemplateName, v.Template)

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
