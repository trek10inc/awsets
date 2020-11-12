package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
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

func (l AWSCloudFormationStack) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	unmapped := make(map[string]int)
	svc := cloudformation.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeStacks(cfg.Context, &cloudformation.DescribeStacksInput{
			NextToken: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.Stacks {
			stackArn := arn.ParseP(v.StackId)
			r := resource.New(cfg, resource.CloudFormationStack, stackArn.ResourceId, v.StackName, v)

			err = Paginator(func(nt2 *string) (*string, error) {
				resourcesRes, err := svc.ListStackResources(cfg.Context, &cloudformation.ListStackResourcesInput{
					StackName: v.StackName,
					NextToken: nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get resources for stack %s: %w", *v.StackName, err)
				}
				for _, rsum := range resourcesRes.StackResourceSummaries {
					rt, err := resource.FromCfn(*rsum.ResourceType)
					if err != nil {
						unmapped[*rsum.ResourceType]++
					}
					if rt == resource.Unnecessary {
						continue
					}
					if rsum.PhysicalResourceId == nil {
						// If stack is in certain statuses (like DELETE_FAILED) the physical id may be nil as this
						// particular resource may have been deleted
						continue
					}
					resourceId := *rsum.PhysicalResourceId
					if strings.Contains(resourceId, "arn:") {
						resourceArn := arn.Parse(resourceId)
						resourceId = resourceArn.ResourceId
					}
					r.AddRelation(rt, resourceId, rsum.ResourceType)
				}

				return resourcesRes.NextToken, nil
			})
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	if len(unmapped) > 0 {
		stacksMsg := fmt.Sprintf("unmapped cf types:\n")
		for k, v := range unmapped {
			stacksMsg += fmt.Sprintf("%s,%03d\n", k, v)
		}
		cfg.SendStatus(option.StatusLogInfo, stacksMsg)
	}
	return rg, err
}
