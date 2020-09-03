package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/resource"
)

type AWSDynamoDBTable struct {
}

func init() {
	i := AWSDynamoDBTable{}
	listers = append(listers, i)
}

func (l AWSDynamoDBTable) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DynamoDbTable}
}

func (l AWSDynamoDBTable) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := dynamodb.New(ctx.AWSCfg)

	req := svc.ListTablesRequest(&dynamodb.ListTablesInput{
		Limit: aws.Int64(100),
	})
	paginator := dynamodb.NewListTablesPaginator(req)
	rg := resource.NewGroup()
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, table := range page.TableNames {

			tableReq := svc.DescribeTableRequest(&dynamodb.DescribeTableInput{
				TableName: aws.String(table),
			})
			res, err := tableReq.Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to describe table %s: %w", table, err)
			}
			tableArn := arn.ParseP(res.Table.TableArn)
			r := resource.New(ctx, resource.DynamoDbTable, tableArn.ResourceId, res.Table.TableName, res.Table)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
