package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dax"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSDAXCluster struct {
}

func init() {
	i := AWSDAXCluster{}
	listers = append(listers, i)
}

func (l AWSDAXCluster) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DAXCluster}
}

func (l AWSDAXCluster) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := dax.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeClusters(ctx.Context, &dax.DescribeClustersInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "Access Denied to API Version: DAX_V3") {
				// Regions that don't support DAX return access denied
				return nil, nil
			}
			return nil, fmt.Errorf("failed to list dax clusters: %w", err)
		}
		for _, v := range res.Clusters {
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
		return res.NextToken, nil
	})
	return rg, err
}
