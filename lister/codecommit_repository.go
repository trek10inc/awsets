package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/codecommit"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSCodeCommitRepository struct {
}

func init() {
	i := AWSCodeCommitRepository{}
	listers = append(listers, i)
}

func (l AWSCodeCommitRepository) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CodeCommitRepository}
}

func (l AWSCodeCommitRepository) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := codecommit.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListRepositories(ctx.Context, &codecommit.ListRepositoriesInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, id := range res.Repositories {
			repo, err := svc.GetRepository(ctx.Context, &codecommit.GetRepositoryInput{
				RepositoryName: id.RepositoryName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get repository %s: %w", *id.RepositoryId, err)
			}
			if v := repo.RepositoryMetadata; v != nil {
				r := resource.New(ctx, resource.CodeCommitRepository, v.RepositoryId, v.RepositoryName, v)
				rg.AddResource(r)
			}
		}
		return res.NextToken, nil
	})
	return rg, err
}
