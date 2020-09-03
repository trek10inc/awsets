package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/dax"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/resource"
)

type AWSDAXCluster struct {
}

func init() {
	i := AWSDAXCluster{}
	listers = append(listers, i)
}

func (l AWSDAXCluster) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CodeBuildProject}
}

func (l AWSDAXCluster) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := dax.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var nextToken *string
	for {
		clusters, err := svc.DescribeClustersRequest(&dax.DescribeClustersInput{
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
			return rg, fmt.Errorf("failed to list dax clusters: %w", err)
		}
		for _, v := range clusters.Clusters {
			clusterArn := arn.ParseP(v.ClusterArn)
			r := resource.New(ctx, resource.DAXCluster, clusterArn.ResourceId, v.ClusterName, v)
			r.AddRelation(resource.DAXParameterGroup, v.ParameterGroup.ParameterGroupName, "")
			r.AddRelation(resource.DAXSubnetGroup, v.SubnetGroup, "")
			for _, sg := range v.SecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, sg.SecurityGroupIdentifier, "")
			}
			r.AddARNRelation(resource.IamRole, v.IamRoleArn)
			rg.AddResource(r)
		}

		if clusters.NextToken == nil {
			break
		}
		nextToken = clusters.NextToken
	}
	return rg, nil
}
