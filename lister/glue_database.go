package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/glue"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSGlueDatabase struct {
}

func init() {
	i := AWSGlueDatabase{}
	listers = append(listers, i)
}

func (l AWSGlueDatabase) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.GlueDatabase, resource.GlueTable}
}

func (l AWSGlueDatabase) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := glue.New(ctx.AWSCfg)
	req := svc.GetDatabasesRequest(&glue.GetDatabasesInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := glue.NewGetDatabasesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.DatabaseList {
			r := resource.New(ctx, resource.GlueDatabase, v.Name, v.Name, v)

			tablesPaginator := glue.NewGetTablesPaginator(svc.GetTablesRequest(&glue.GetTablesInput{
				DatabaseName: v.Name,
				MaxResults:   aws.Int64(100),
			}))
			for tablesPaginator.Next(ctx.Context) {
				tablesPage := tablesPaginator.CurrentPage()
				for _, table := range tablesPage.TableList {
					tableR := resource.New(ctx, resource.GlueTable, makeGlueTableId(v.Name, table.Name), table.Name, table)
					tableR.AddRelation(resource.GlueDatabase, v.Name, "")
					rg.AddResource(tableR)
				}
			}
			err := tablesPaginator.Err()
			if err != nil {
				return rg, fmt.Errorf("failed to list tables for database %s: %w", *v.Name, err)
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}

func makeGlueTableId(dbName, tableName *string) string {
	return *dbName + "-" + *tableName
}
