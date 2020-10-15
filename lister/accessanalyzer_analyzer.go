package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSAccessAnalyzerAnalyzer struct {
}

func init() {
	i := AWSAccessAnalyzerAnalyzer{}
	listers = append(listers, i)
}

func (l AWSAccessAnalyzerAnalyzer) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.AccessAnalyzerAnalyzer}
}

func (l AWSAccessAnalyzerAnalyzer) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := accessanalyzer.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		req, err := svc.ListAnalyzers(cfg.Context, &accessanalyzer.ListAnalyzersInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range req.Analyzers {
			r := resource.New(cfg, resource.AccessAnalyzerAnalyzer, v.Name, v.Name, v)
			rg.AddResource(r)
		}
		return req.NextToken, nil
	})
	return rg, err
}
