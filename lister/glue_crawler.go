package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
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

func (l AWSGlueCrawler) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := glue.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.GetCrawlers(cfg.Context, &glue.GetCrawlersInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.Crawlers {
			r := resource.NewVersion(cfg, resource.GlueCrawler, v.Name, v.Name, v.Version, v)
			r.AddARNRelation(resource.IamRole, v.Role)
			r.AddRelation(resource.GlueDatabase, v.DatabaseName, "")
			// TODO: review relationships to s3, ddb, jdbc

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
