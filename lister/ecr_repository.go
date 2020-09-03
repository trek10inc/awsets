package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/trek10inc/awsets/arn"
)

type AWSEcrRepository struct {
}

func init() {
	i := AWSEcrRepository{}
	listers = append(listers, i)
}

func (l AWSEcrRepository) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.EcrRepository}
}

func (l AWSEcrRepository) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ecr.New(ctx.AWSCfg)

	req := svc.DescribeRepositoriesRequest(&ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int64(1000),
	})

	paginator := ecr.NewDescribeRepositoriesPaginator(req)
	rg := resource.NewGroup()
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, repo := range page.Repositories {
			repoArn := arn.ParseP(repo.RepositoryArn)
			r := resource.New(ctx, resource.EcrRepository, repoArn.ResourceId, repo.RepositoryName, repo)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
