package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/dax"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"
)

type AWSDAXSubnetGroup struct {
}

func init() {
	i := AWSDAXSubnetGroup{}
	listers = append(listers, i)
}

func (l AWSDAXSubnetGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DAXSubnetGroup}
}

func (l AWSDAXSubnetGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := dax.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		groups, err := svc.DescribeSubnetGroupsRequest(&dax.DescribeSubnetGroupsInput{
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
			return rg, fmt.Errorf("failed to list dax subnet groups: %w", err)
		}
		for _, v := range groups.SubnetGroups {
			r := resource.New(ctx, resource.DAXSubnetGroup, v.SubnetGroupName, v.SubnetGroupName, v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			for _, sn := range v.Subnets {
				r.AddRelation(resource.Ec2Subnet, sn.SubnetIdentifier, "")
			}
			rg.AddResource(r)
		}

		if groups.NextToken == nil {
			break
		}
		nextToken = groups.NextToken
	}
	return rg, nil
}
