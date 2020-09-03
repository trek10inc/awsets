package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSAthenaWorkGroup struct {
}

func init() {
	i := AWSAthenaWorkGroup{}
	listers = append(listers, i)
}

func (l AWSAthenaWorkGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.AthenaWorkGroup, resource.AthenaNamedQuery}
}

func (l AWSAthenaWorkGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := athena.New(ctx.AWSCfg)

	rg := resource.NewGroup()

	req := svc.ListWorkGroupsRequest(&athena.ListWorkGroupsInput{
		MaxResults: aws.Int64(50),
	})

	paginator := athena.NewListWorkGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, wg := range page.WorkGroups {
			r := resource.New(ctx, resource.AthenaWorkGroup, wg.Name, wg.Name, wg)
			nqReq := svc.ListNamedQueriesRequest(&athena.ListNamedQueriesInput{
				MaxResults: aws.Int64(50),
				WorkGroup:  wg.Name,
			})
			nqPaginator := athena.NewListNamedQueriesPaginator(nqReq)
			for nqPaginator.Next(ctx.Context) {
				nqPage := nqPaginator.CurrentPage()
				for _, id := range nqPage.NamedQueryIds {
					query, err := svc.GetNamedQueryRequest(&athena.GetNamedQueryInput{
						NamedQueryId: &id,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to get named query %s: %w", id, err)
					}
					if v := query.NamedQuery; v != nil {
						nqR := resource.New(ctx, resource.AthenaNamedQuery, v.NamedQueryId, v.Name, v)
						nqR.AddRelation(resource.AthenaWorkGroup, wg.Name, "")
						rg.AddResource(nqR)
					}
				}
			}
			err := nqPaginator.Err()
			if err != nil {
				return rg, fmt.Errorf("failed to list named querys for workgroup %s: %w", *wg.Name, err)
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
