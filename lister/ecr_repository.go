package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := ecr.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeRepositories(ctx.Context, &ecr.DescribeRepositoriesInput{
			MaxResults: aws.Int32(1000),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, repo := range res.Repositories {
			repoArn := arn.ParseP(repo.RepositoryArn)
			r := resource.New(ctx, resource.EcrRepository, repoArn.ResourceId, repo.RepositoryName, repo)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
