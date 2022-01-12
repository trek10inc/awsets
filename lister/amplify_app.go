package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/amplify"
	"github.com/trek10inc/awsets/arn"
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

	svc := amplify.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		apps, err := svc.ListApps(ctx.Context, &amplify.ListAppsInput{
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list amplify apps: %w", err)
		}
		for _, v := range apps.Apps {
			r := resource.New(ctx, resource.AmplifyApp, v.AppId, v.Name, v)

			// add Amplify Branches
			err = Paginator(func(nt2 *string) (*string, error) {
				branches, err := svc.ListBranches(ctx.Context, &amplify.ListBranchesInput{
					AppId:      v.AppId,
					MaxResults: 50,
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list branches for app %s: %w", *v.AppId, err)
				}
				for _, branch := range branches.Branches {
					branchArn := arn.ParseP(branch.BranchArn)
					branchR := resource.New(ctx, resource.AmplifyBranch, branchArn.ResourceId, branch.DisplayName, branch)
					branchR.AddRelation(resource.AmplifyApp, v.AppId, "")
					rg.AddResource(branchR)
				}
				return branches.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// add Amplify Domains
			err = Paginator(func(nt2 *string) (*string, error) {
				domains, err := svc.ListDomainAssociations(ctx.Context, &amplify.ListDomainAssociationsInput{
					AppId:      v.AppId,
					MaxResults: 50,
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list domains for app %s: %w", *v.AppId, err)
				}
				for _, domain := range domains.DomainAssociations {
					domainR := resource.New(ctx, resource.AmplifyDomain, domain.DomainName, domain.DomainName, domain)
					domainR.AddRelation(resource.AmplifyApp, v.AppId, "")
					rg.AddResource(domainR)
				}
				return domains.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			rg.AddResource(r)
		}
		return apps.NextToken, nil
	})
	return rg, err
}
