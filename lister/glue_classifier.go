package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSGlueClassifier struct {
}

func init() {
	i := AWSGlueClassifier{}
	listers = append(listers, i)
}

func (l AWSGlueClassifier) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GlueClassifier,
	}
}

func (l AWSGlueClassifier) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := glue.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.GetClassifiers(cfg.Context, &glue.GetClassifiersInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.Classifiers {

			if x := v.CsvClassifier; x != nil {
				r := resource.NewVersion(cfg, resource.GlueClassifier, x.Name, x.Name, x.Version, x)
				rg.AddResource(r)
			} else if x := v.GrokClassifier; x != nil {
				r := resource.NewVersion(cfg, resource.GlueClassifier, x.Name, x.Name, x.Version, x)
				rg.AddResource(r)
			} else if x := v.JsonClassifier; x != nil {
				r := resource.NewVersion(cfg, resource.GlueClassifier, x.Name, x.Name, x.Version, x)
				rg.AddResource(r)
			} else if x := v.XMLClassifier; x != nil {
				r := resource.NewVersion(cfg, resource.GlueClassifier, x.Name, x.Name, x.Version, x)
				rg.AddResource(r)
			}
		}
		return res.NextToken, nil
	})
	return rg, err
}
