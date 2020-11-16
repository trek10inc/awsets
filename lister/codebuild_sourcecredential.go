package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSCodebuildSourceCredential struct {
}

func init() {
	i := AWSCodebuildSourceCredential{}
	listers = append(listers, i)
}

func (l AWSCodebuildSourceCredential) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.CodeBuildSourceCredential,
	}
}

func (l AWSCodebuildSourceCredential) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := codebuild.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	res, err := svc.ListSourceCredentials(ctx.Context, &codebuild.ListSourceCredentialsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list codebuild source credentials: %w", err)
	}
	for _, sc := range res.SourceCredentialsInfos {
		credArn := arn.ParseP(sc.Arn)
		r := resource.New(ctx, resource.CodeBuildProject, credArn.ResourceId, credArn.ResourceType, sc)
		rg.AddResource(r)
	}
	return rg, err
}
