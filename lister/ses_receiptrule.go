package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSSESReceiptRule struct {
}

func init() {
	i := AWSSESReceiptRule{}
	listers = append(listers, i)
}

func (l AWSSESReceiptRule) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.SesReceiptRule,
		resource.SesReceiptRuleSet,
	}
}

func (l AWSSESReceiptRule) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ses.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListReceiptRuleSets(cfg.Context, &ses.ListReceiptRuleSetsInput{
			NextToken: nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "Unavailable Operation") {
				// If SES isn't supported in a region, returns Unavailable Operation error
				return nil, nil
			}
			return nil, err
		}
		for _, rs := range res.RuleSets {
			v, err := svc.DescribeReceiptRuleSet(cfg.Context, &ses.DescribeReceiptRuleSetInput{
				RuleSetName: rs.Name,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get receipt rule set %s: %w", *rs.Name, err)
			}
			r := resource.New(cfg, resource.SesReceiptRuleSet, v.Metadata.Name, v.Metadata.Name, v.Metadata)

			for _, rule := range v.Rules {
				rr := resource.New(cfg, resource.SesReceiptRule, rule.Name, rule.Name, rule)
				rr.AddRelation(resource.SesReceiptRuleSet, v.Metadata.Name, "")
				for _, action := range rule.Actions {
					if action.BounceAction != nil {
						rr.AddARNRelation(resource.SnsTopic, action.BounceAction.TopicArn)
					}
					if action.LambdaAction != nil {
						rr.AddARNRelation(resource.SnsTopic, action.LambdaAction.TopicArn)
						rr.AddARNRelation(resource.LambdaFunction, action.LambdaAction.FunctionArn)
					}
					if action.S3Action != nil {
						rr.AddRelation(resource.S3Bucket, action.S3Action.BucketName, "")
						rr.AddARNRelation(resource.KmsKey, action.S3Action.KmsKeyArn)
					}
					if action.SNSAction != nil {
						rr.AddARNRelation(resource.SnsTopic, action.SNSAction.TopicArn)
					}
				}
				rg.AddResource(r)
			}

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
