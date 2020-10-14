package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qldb"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSQLDBLedger struct {
}

func init() {
	i := AWSQLDBLedger{}
	listers = append(listers, i)
}

func (l AWSQLDBLedger) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.QLDBLedger}
}

func (l AWSQLDBLedger) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := qldb.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListLedgers(ctx.Context, &qldb.ListLedgersInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.Ledgers {
			r := resource.New(ctx, resource.QLDBLedger, v.Name, v.Name, v)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
