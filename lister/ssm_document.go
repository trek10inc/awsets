package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSSsmDocument struct {
}

func init() {
	i := AWSSsmDocument{}
	listers = append(listers, i)
}

func (l AWSSsmDocument) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.SsmDocument}
}

func (l AWSSsmDocument) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ssm.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListDocuments(ctx.Context, &ssm.ListDocumentsInput{
			MaxResults: aws.Int32(50),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, docId := range res.DocumentIdentifiers {
			if docId.Owner != nil && *docId.Owner != "Amazon" { // TODO: should Amazon things be filtered?
				r := resource.New(ctx, resource.SsmDocument, docId.Name, docId.Name, docId)
				rg.AddResource(r)
			}
		}
		return res.NextToken, nil
	})
	return rg, err
}
