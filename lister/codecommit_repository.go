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

	svc := codecommit.New(ctx.AWSCfg)
	rg := resource.NewGroup()

	req := svc.ListRepositoriesRequest(&codecommit.ListRepositoriesInput{})

	paginator := codecommit.NewListRepositoriesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, id := range page.Repositories {
			repo, err := svc.GetRepositoryRequest(&codecommit.GetRepositoryInput{
				RepositoryName: id.RepositoryName,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to get repository %s: %w", *id.RepositoryId, err)
			}
			if v := repo.RepositoryMetadata; v != nil {
				r := resource.New(ctx, resource.CodeCommitRepository, v.RepositoryId, v.RepositoryName, v)
				rg.AddResource(r)
			}
		}
	}
	err := paginator.Err()
	return rg, err
}
