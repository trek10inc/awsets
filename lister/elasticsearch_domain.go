package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSElasticsearchDomain struct {
}

func init() {
	i := AWSElasticsearchDomain{}
	listers = append(listers, i)
}

func (l AWSElasticsearchDomain) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ElasticsearchDomain}
}

func (l AWSElasticsearchDomain) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := elasticsearchservice.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	domainListRes, err := svc.ListDomainNames(ctx.Context, &elasticsearchservice.ListDomainNamesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list domain names: %w", err)
	}

	domainsByFive := make([][]string, 0)
	domainNames := make([]string, 0)
	for _, domain := range domainListRes.DomainNames {
		domainNames = append(domainNames, *domain.DomainName)
		if len(domainNames) == 5 {
			domainsByFive = append(domainsByFive, domainNames)
			domainNames = make([]string, 0)
		}
	}
	if len(domainNames) > 0 {
		domainsByFive = append(domainsByFive, domainNames)
	}

	for _, fiveDomains := range domainsByFive {
		domains, err := svc.DescribeElasticsearchDomains(ctx.Context, &elasticsearchservice.DescribeElasticsearchDomainsInput{
			DomainNames: fiveDomains, // max of 5 :(
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe domains: %w", err)
		}

		for _, domain := range domains.DomainStatusList {
			domainArn := arn.ParseP(domain.ARN)
			r := resource.New(ctx, resource.ElasticsearchDomain, domainArn.ResourceId, domain.DomainName, domain)

			if domain.VPCOptions != nil {
				r.AddRelation(resource.Ec2Vpc, domain.VPCOptions.VPCId, "")
				for _, sg := range domain.VPCOptions.SecurityGroupIds {
					r.AddRelation(resource.Ec2SecurityGroup, sg, "")
				}
			}
			if domain.EncryptionAtRestOptions != nil {
				r.AddARNRelation(resource.KmsKey, domain.EncryptionAtRestOptions.KmsKeyId)
			}
			tagsRes, err := svc.ListTags(ctx.Context, &elasticsearchservice.ListTagsInput{
				ARN: domain.ARN,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list tags for %s: %w", *domain.DomainName, err)
			}
			for _, tag := range tagsRes.TagList {
				r.Tags[*tag.Key] = *tag.Value
			}
			//TODO relationship to cognito user pool
			rg.AddResource(r)
		}
	}

	return rg, nil
}
