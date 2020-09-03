package lister

import (
	"strings"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/resource"
)

type AWSCloudFormationStack struct {
}

func init() {
	i := AWSCloudFormationStack{}
	listers = append(listers, i)
}

func (l AWSCloudFormationStack) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CloudFormationStack}
}

func (l AWSCloudFormationStack) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	unmapped := make(map[string]int)
	svc := cloudformation.New(ctx.AWSCfg)

	req := svc.DescribeStacksRequest(&cloudformation.DescribeStacksInput{})

	rg := resource.NewGroup()
	paginator := cloudformation.NewDescribeStacksPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Stacks {
			stackArn := arn.ParseP(v.StackId)
			r := resource.New(ctx, resource.CloudFormationStack, stackArn.ResourceId, v.StackName, v)

			rreq := svc.ListStackResourcesRequest(&cloudformation.ListStackResourcesInput{
				StackName: v.StackName,
			})
			rpaginator := cloudformation.NewListStackResourcesPaginator(rreq)
			for rpaginator.Next(ctx.Context) {
				rpage := rpaginator.CurrentPage()
				for _, rsum := range rpage.StackResourceSummaries {
					rt, err := resource.FromCfn(aws.StringValue(rsum.ResourceType))
					if err != nil {
						unmapped[*rsum.ResourceType]++
					}
					if rt == resource.Unnecessary {
						continue
					}
					resourceId := aws.StringValue(rsum.PhysicalResourceId)
					if strings.Contains(resourceId, "arn:") {
						resourceArn := arn.Parse(resourceId)
						resourceId = resourceArn.ResourceId
					}
					r.AddRelation(rt, resourceId, rsum.ResourceType)
				}
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	if len(unmapped) > 0 {
		ctx.Logger.Errorf("unmapped cf types for region %s:\n", ctx.Region())
		for k, v := range unmapped {
			ctx.Logger.Errorf("%s,%03d\n", k, v)
		}
	}
	return rg, err
}
