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
	svc := accessanalyzer.New(ctx.AWSCfg)

	req := svc.ListAnalyzersRequest(&accessanalyzer.ListAnalyzersInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := accessanalyzer.NewListAnalyzersPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Analyzers {
			r := resource.New(ctx, resource.AccessAnalyzerAnalyzer, v.Name, v.Name, v)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
