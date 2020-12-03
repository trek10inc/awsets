package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSEksCluster struct {
}

func init() {
	i := AWSEksCluster{}
	listers = append(listers, i)
}

func (l AWSEksCluster) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.EksCluster,
		resource.EksNodeGroup,
		resource.EksFargateProfile,
	}
}

func (l AWSEksCluster) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := eks.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListClusters(ctx.Context, &eks.ListClustersInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf("Account %s is not authorized to use this service", ctx.AccountId)) {
				// If EKS is not supported in a region, returns access denied
				return nil, nil
			}
			return nil, err
		}
		if len(res.Clusters) == 0 {
			return nil, nil
		}
		for _, clusterName := range res.Clusters {
			clusterRes, err := svc.DescribeCluster(ctx.Context, &eks.DescribeClusterInput{
				Name: &clusterName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe cluster %s: %w", clusterName, err)
			}
			cluster := clusterRes.Cluster
			r := resource.New(ctx, resource.EksCluster, cluster.Name, cluster.Name, cluster)
			r.AddARNRelation(resource.IamRole, cluster.RoleArn)

			// Node groups
			err = Paginator(func(nt2 *string) (*string, error) {
				nodeGroups, err := svc.ListNodegroups(ctx.Context, &eks.ListNodegroupsInput{
					ClusterName: &clusterName,
					MaxResults:  aws.Int32(100),
					NextToken:   nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list node groups for cluster %s: %w", clusterName, err)
				}
				for _, ngName := range nodeGroups.Nodegroups {
					ngRes, err := svc.DescribeNodegroup(ctx.Context, &eks.DescribeNodegroupInput{
						ClusterName:   &clusterName,
						NodegroupName: &ngName,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to describe node group %s for cluster %s: %w", ngName, clusterName, err)
					}
					ng := ngRes.Nodegroup
					ngArn := arn.ParseP(ng.NodegroupArn)
					ngResource := resource.New(ctx, resource.EksNodeGroup, ngArn.ResourceId, ng.NodegroupName, ng)
					ngResource.AddRelation(resource.EksCluster, clusterName, "")
					for _, sn := range ng.Subnets {
						ngResource.AddRelation(resource.Ec2Subnet, sn, "")
					}
					rg.AddResource(ngResource)
				}
				return nodeGroups.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Fargate profiles
			err = Paginator(func(nt2 *string) (*string, error) {
				profiles, err := svc.ListFargateProfiles(ctx.Context, &eks.ListFargateProfilesInput{
					ClusterName: &clusterName,
					NextToken:   nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list fargate profiles for cluster %s: %w", clusterName, err)
				}
				for _, fpName := range profiles.FargateProfileNames {
					fpRes, err := svc.DescribeFargateProfile(ctx.Context, &eks.DescribeFargateProfileInput{
						ClusterName:        &clusterName,
						FargateProfileName: &fpName,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to describe fargate profile %s for cluster %s: %w", fpName, clusterName, err)
					}
					if fp := fpRes.FargateProfile; fp != nil {
						fpResource := resource.New(ctx, resource.EksFargateProfile, fmt.Sprintf("%s-%s", clusterName, fpName), fp.FargateProfileName, fp)
						for _, sn := range fp.Subnets {
							fpResource.AddRelation(resource.Ec2Subnet, sn, "")
						}
						fpResource.AddARNRelation(resource.IamRole, fp.PodExecutionRoleArn)
						fpResource.AddRelation(resource.EksCluster, clusterName, "")
						rg.AddResource(fpResource)
					}
				}
				return profiles.NextToken, nil
			})
			if err != nil {
				return nil, err
			}
			rg.AddResource(r)
		}

		return res.NextToken, nil
	})
	return rg, err
}
