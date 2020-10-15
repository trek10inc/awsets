package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
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

func (l AWSEksCluster) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := eks.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListClusters(cfg.Context, &eks.ListClustersInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf("Account %s is not authorized to use this service", cfg.AccountId)) {
				// If EKS is not supported in a region, returns access denied
				return nil, nil
			}
			return nil, err
		}
		if len(res.Clusters) == 0 {
			return nil, nil
		}
		for _, clusterName := range res.Clusters {
			clusterRes, err := svc.DescribeCluster(cfg.Context, &eks.DescribeClusterInput{
				Name: clusterName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe cluster %s: %w", *clusterName, err)
			}
			cluster := clusterRes.Cluster
			r := resource.New(cfg, resource.EksCluster, cluster.Name, cluster.Name, cluster)
			r.AddARNRelation(resource.IamRole, cluster.RoleArn)

			// Node groups
			err = Paginator(func(nt2 *string) (*string, error) {
				nodeGroups, err := svc.ListNodegroups(cfg.Context, &eks.ListNodegroupsInput{
					ClusterName: clusterName,
					MaxResults:  aws.Int32(100),
					NextToken:   nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list node groups for cluster %s: %w", *clusterName, err)
				}
				for _, ngName := range nodeGroups.Nodegroups {
					ngRes, err := svc.DescribeNodegroup(cfg.Context, &eks.DescribeNodegroupInput{
						ClusterName:   clusterName,
						NodegroupName: ngName,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to describe node group %s for cluster %s: %w", *ngName, *clusterName, err)
					}
					ng := ngRes.Nodegroup
					ngArn := arn.ParseP(ng.NodegroupArn)
					ngResource := resource.New(cfg, resource.EksNodeGroup, ngArn.ResourceId, ng.NodegroupName, ng)
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
				profiles, err := svc.ListFargateProfiles(cfg.Context, &eks.ListFargateProfilesInput{
					ClusterName: clusterName,
					NextToken:   nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list fargate profiles for cluster %s: %w", *clusterName, err)
				}
				for _, fpName := range profiles.FargateProfileNames {
					fpRes, err := svc.DescribeFargateProfile(cfg.Context, &eks.DescribeFargateProfileInput{
						ClusterName:        clusterName,
						FargateProfileName: fpName,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to describe fargate profile %s for cluster %s: %w", *fpName, *clusterName, err)
					}
					if fp := fpRes.FargateProfile; fp != nil {
						fpResource := resource.New(cfg, resource.EksFargateProfile, fmt.Sprintf("%s-%s", *clusterName, *fpName), fp.FargateProfileName, fp)
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
