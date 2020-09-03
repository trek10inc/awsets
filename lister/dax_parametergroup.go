package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"

	"github.com/aws/aws-sdk-go-v2/service/dax"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"
)

type AWSDAXParameterGroup struct {
}

func init() {
	i := AWSDAXParameterGroup{}
	listers = append(listers, i)
}

func (l AWSDAXParameterGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DAXParameterGroup}
}

func (l AWSDAXParameterGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := dax.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		groups, err := svc.DescribeParameterGroupsRequest(&dax.DescribeParameterGroupsInput{
			MaxResults: aws.Int64(100),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				if aerr.Code() == dax.ErrCodeInvalidParameterValueException &&
					strings.Contains(aerr.Message(), "Access Denied to API Version: DAX_V3") {
					// Regions that don't support DAX return access denied
					return rg, nil
				}
			}
			return rg, fmt.Errorf("failed to list dax parameter groups: %w", err)
		}
		for _, v := range groups.ParameterGroups {
			r := resource.New(ctx, resource.DAXParameterGroup, v.ParameterGroupName, v.ParameterGroupName, v)
			rg.AddResource(r)
		}

		if groups.NextToken == nil {
			break
		}
		nextToken = groups.NextToken
	}
	return rg, nil
}
