package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/service/amplify"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"
)

type AWSAmplifyApp struct {
}

func init() {
	i := AWSAmplifyApp{}
	listers = append(listers, i)
}

func (l AWSAmplifyApp) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.AmplifyApp,
		resource.AmplifyBranch,
		resource.AmplifyDomain,
	}
}

func (l AWSAmplifyApp) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := amplify.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		apps, err := svc.ListAppsRequest(&amplify.ListAppsInput{
			MaxResults: aws.Int64(100),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list amplify apps: %w", err)
		}
		for _, v := range apps.Apps {
			r := resource.New(ctx, resource.AmplifyApp, v.AppId, v.Name, v)

			// add Amplify Branches
			var branchNextToken *string
			for {
				branches, err := svc.ListBranchesRequest(&amplify.ListBranchesInput{
					AppId:      v.AppId,
					MaxResults: aws.Int64(50),
					NextToken:  branchNextToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list branches for app %s: %w", *v.AppId, err)
				}
				for _, branch := range branches.Branches {
					branchArn := arn.ParseP(branch.BranchArn)
					branchR := resource.New(ctx, resource.AmplifyBranch, branchArn.ResourceId, branch.DisplayName, branch)
					branchR.AddRelation(resource.AmplifyApp, v.AppId, "")
					rg.AddResource(branchR)
				}
				if branches.NextToken == nil {
					break
				}
				branchNextToken = branches.NextToken
			}

			// add Amplify Domains
			var domainNextToken *string
			for {
				domains, err := svc.ListDomainAssociationsRequest(&amplify.ListDomainAssociationsInput{
					AppId:      v.AppId,
					MaxResults: aws.Int64(50),
					NextToken:  domainNextToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list domains for app %s: %w", *v.AppId, err)
				}
				for _, domain := range domains.DomainAssociations {
					domainR := resource.New(ctx, resource.AmplifyDomain, domain.DomainName, domain.DomainName, domain)
					domainR.AddRelation(resource.AmplifyApp, v.AppId, "")
					rg.AddResource(domainR)
				}
				if domains.NextToken == nil {
					break
				}
				domainNextToken = domains.NextToken
			}

			rg.AddResource(r)
		}

		if apps.NextToken == nil {
			break
		}
		nextToken = apps.NextToken
	}
	return rg, nil
}
