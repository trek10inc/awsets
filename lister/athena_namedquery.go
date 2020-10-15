package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/trek10inc/awsets/option"
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

func (l AWSAthenaWorkGroup) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := athena.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListWorkGroups(cfg.Context, &athena.ListWorkGroupsInput{
			MaxResults: aws.Int32(50),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, wg := range res.WorkGroups {
			r := resource.New(cfg, resource.AthenaWorkGroup, wg.Name, wg.Name, wg)

			err = Paginator(func(nt2 *string) (*string, error) {
				nqRes, err := svc.ListNamedQueries(cfg.Context, &athena.ListNamedQueriesInput{
					MaxResults: aws.Int32(50),
					WorkGroup:  wg.Name,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list named querys for workgroup %s: %w", *wg.Name, err)
				}
				for _, id := range nqRes.NamedQueryIds {
					query, err := svc.GetNamedQuery(cfg.Context, &athena.GetNamedQueryInput{
						NamedQueryId: id,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to get named query %s: %w", *id, err)
					}
					if v := query.NamedQuery; v != nil {
						nqR := resource.New(cfg, resource.AthenaNamedQuery, v.NamedQueryId, v.Name, v)
						nqR.AddRelation(resource.AthenaWorkGroup, wg.Name, "")
						rg.AddResource(nqR)
					}
				}
				return nqRes.NextToken, nil
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
