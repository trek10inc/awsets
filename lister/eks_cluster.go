package lister

import (
	"fmt"
	"strings"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/trek10inc/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	svc := eks.New(ctx.AWSCfg)
	req := svc.ListClustersRequest(&eks.ListClustersInput{
		MaxResults: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := eks.NewListClustersPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		if len(page.Clusters) == 0 {
			continue
		}
		for _, clusterName := range page.Clusters {
			clusterRes, err := svc.DescribeClusterRequest(&eks.DescribeClusterInput{
				Name: &clusterName,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to describe cluster %s: %w", clusterName, err)
			}
			cluster := clusterRes.Cluster
			r := resource.New(ctx, resource.EksCluster, cluster.Name, cluster.Name, cluster)
			r.AddARNRelation(resource.IamRole, cluster.RoleArn)

			// Node groups
			ngPaginator := eks.NewListNodegroupsPaginator(svc.ListNodegroupsRequest(&eks.ListNodegroupsInput{
				ClusterName: &clusterName,
				MaxResults:  aws.Int64(100),
			}))
			for ngPaginator.Next(ctx.Context) {
				ngPage := ngPaginator.CurrentPage()
				for _, ngName := range ngPage.Nodegroups {
					ngRes, err := svc.DescribeNodegroupRequest(&eks.DescribeNodegroupInput{
						ClusterName:   &clusterName,
						NodegroupName: &ngName,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to describe node group %s for cluster %s: %w", ngName, clusterName, err)
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
			}
			err = ngPaginator.Err()
			if err != nil {
				return rg, fmt.Errorf("failed to list node groups for cluster %s: %w", clusterName, err)
			}

			// Fargate profiles
			fpPaginator := eks.NewListFargateProfilesPaginator(svc.ListFargateProfilesRequest(&eks.ListFargateProfilesInput{
				ClusterName: &clusterName,
			}))
			for fpPaginator.Next(ctx.Context) {
				fpPage := fpPaginator.CurrentPage()
				for _, fpName := range fpPage.FargateProfileNames {
					fpRes, err := svc.DescribeFargateProfileRequest(&eks.DescribeFargateProfileInput{
						ClusterName:        &clusterName,
						FargateProfileName: &fpName,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to describe fargate profile %s for cluster %s: %w", fpName, clusterName, err)
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
			}

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "AccessDeniedException" &&
				strings.Contains(aerr.Message(), fmt.Sprintf("Account %s is not authorized to use this service", ctx.AccountId)) {
				// If EKS is not supported in a region, returns access denied
				err = nil
			}
		}
	}
	return rg, err
}
