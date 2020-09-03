package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
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
	svc := ssm.New(ctx.AWSCfg)
	req := svc.ListDocumentsRequest(&ssm.ListDocumentsInput{
		MaxResults: aws.Int64(50),
	})

	rg := resource.NewGroup()
	paginator := ssm.NewListDocumentsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, docId := range page.DocumentIdentifiers {
			if docId.Owner != nil && *docId.Owner != "Amazon" { // TODO: should Amazon things be filtered?
				r := resource.New(ctx, resource.SsmDocument, docId.Name, docId.Name, docId)
				rg.AddResource(r)
			}
		}
	}
	err := paginator.Err()
	return rg, err
}
