package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/trek10inc/awsets/resource"
)

type AWSCloudFormationStackSet struct {
}

func init() {
	i := AWSCloudFormationStackSet{}
	listers = append(listers, i)
}

func (l AWSCloudFormationStackSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CloudFormationStackSet}
}

func (l AWSCloudFormationStackSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloudformation.New(ctx.AWSCfg)

	req := svc.ListStackSetsRequest(&cloudformation.ListStackSetsInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := cloudformation.NewListStackSetsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, summary := range page.Summaries {
			v, err := svc.DescribeStackSetRequest(&cloudformation.DescribeStackSetInput{
				StackSetName: summary.StackSetName,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to describe stack set %s: %w", *summary.StackSetName, err)
			}
			r := resource.New(ctx, resource.CloudFormationStackSet, v.StackSet.StackSetId, v.StackSet.StackSetName, v.StackSet)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "ValidationException" &&
				strings.Contains(aerr.Message(), "is not supported in this region") {
				// If stacksets are not supported in a region, returns validation exception
				err = nil
			}
		}
	}
	return rg, err
}
