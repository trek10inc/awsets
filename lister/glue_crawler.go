package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/glue"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSGlueCrawler struct {
}

func init() {
	i := AWSGlueCrawler{}
	listers = append(listers, i)
}

func (l AWSGlueCrawler) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GlueCrawler,
	}
}

func (l AWSGlueCrawler) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := glue.NewFromConfig(ctx.AWSCfg)
	res, err := svc.GetCrawlers(ctx.Context, &glue.GetCrawlersInput{
		MaxResults: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := glue.NewGetCrawlersPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Crawlers {
			r := resource.NewVersion(ctx, resource.GlueCrawler, v.Name, v.Name, v.Version, v)
			r.AddARNRelation(resource.IamRole, v.Role)
			r.AddRelation(resource.GlueDatabase, v.DatabaseName, "")
			// TODO: review relationships to s3, ddb, jdbc

			rg.AddResource(r)
		}
	}

	err := paginator.Err()
	return rg, err
}
