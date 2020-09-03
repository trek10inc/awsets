package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"

	"github.com/trek10inc/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/service/redshift"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWSRedshiftCluster struct {
}

func init() {
	i := AWSRedshiftCluster{}
	listers = append(listers, i)
}

func (l AWSRedshiftCluster) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.RedshiftCluster}
}

func (l AWSRedshiftCluster) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := redshift.New(ctx.AWSCfg)

	req := svc.DescribeClustersRequest(&redshift.DescribeClustersInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := redshift.NewDescribeClustersPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, cluster := range page.Clusters {
			r := resource.New(ctx, resource.RedshiftCluster, cluster.ClusterIdentifier, cluster.ClusterIdentifier, cluster)
			r.AddRelation(resource.Ec2Vpc, cluster.VpcId, "")
			r.AddRelation(resource.RedshiftSubnetGroup, cluster.ClusterSubnetGroupName, "")
			for _, sg := range cluster.ClusterSecurityGroups {
				r.AddRelation(resource.RedshiftSecurityGroup, sg.ClusterSecurityGroupName, "")
			}
			for _, sg := range cluster.VpcSecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, sg.VpcSecurityGroupId, "")
			}
			for _, role := range cluster.IamRoles {
				roleArn := arn.ParseP(role.IamRoleArn)
				r.AddRelation(resource.IamRole, roleArn.ResourceId, "")
			}
			r.AddRelation(resource.KmsKey, cluster.KmsKeyId, "")

			if cluster.ElasticIpStatus != nil {
				r.AddRelation(resource.Ec2Eip, cluster.ElasticIpStatus.ElasticIp, "")
			}

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
