package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSDynamoDBTable struct {
}

func init() {
	i := AWSDynamoDBTable{}
	listers = append(listers, i)
}

func (l AWSDynamoDBTable) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.DynamoDbTable,
	}
}

func (l AWSDynamoDBTable) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := dynamodb.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListTables(cfg.Context, &dynamodb.ListTablesInput{
			Limit:                   aws.Int32(100),
			ExclusiveStartTableName: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, table := range res.TableNames {
			tableRes, err := svc.DescribeTable(cfg.Context, &dynamodb.DescribeTableInput{
				TableName: table,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe table %s: %w", *table, err)
			}
			r := resource.New(cfg, resource.DynamoDbTable, tableRes.Table.TableId, tableRes.Table.TableName, tableRes.Table)
			rg.AddResource(r)
		}
		return res.LastEvaluatedTableName, nil
	})
	return rg, err
}
