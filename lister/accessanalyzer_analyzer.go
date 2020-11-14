package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
	"github.com/trek10inc/awsets/context"
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

func (l AWSAccessAnalyzerAnalyzer) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := accessanalyzer.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		req, err := svc.ListAnalyzers(ctx.Context, &accessanalyzer.ListAnalyzersInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range req.Analyzers {
			r := resource.New(ctx, resource.AccessAnalyzerAnalyzer, v.Name, v.Name, v)
			rg.AddResource(r)
		}
		return req.NextToken, nil
	})
	return rg, err
}
