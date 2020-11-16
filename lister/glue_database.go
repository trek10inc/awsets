package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSGlueDatabase struct {
}

func init() {
	i := AWSGlueDatabase{}
	listers = append(listers, i)
}

func (l AWSGlueDatabase) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GlueDatabase,
		resource.GlueTable,
	}
}

func (l AWSGlueDatabase) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := glue.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.GetDatabases(ctx.Context, &glue.GetDatabasesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.DatabaseList {
			r := resource.New(ctx, resource.GlueDatabase, v.Name, v.Name, v)

			// Glue Tables
			err = Paginator(func(nt2 *string) (*string, error) {
				tables, err := svc.GetTables(ctx.Context, &glue.GetTablesInput{
					DatabaseName: v.Name,
					MaxResults:   aws.Int32(100),
					NextToken:    nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list tables for database %s: %w", *v.Name, err)
				}
				for _, table := range tables.TableList {
					tableR := resource.New(ctx, resource.GlueTable, makeGlueTableId(v.Name, table.Name), table.Name, table)
					tableR.AddRelation(resource.GlueDatabase, v.Name, "")
					rg.AddResource(tableR)
				}
				return tables.NextToken, nil
			})
			if err != nil {
				return nil, err
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}

func makeGlueTableId(dbName, tableName *string) string {
	return *dbName + "-" + *tableName
}
