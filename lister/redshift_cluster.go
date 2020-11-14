package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
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
	svc := redshift.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeClusters(ctx.Context, &redshift.DescribeClustersInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			return nil, err
		}
		for _, cluster := range res.Clusters {
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
				r.AddARNRelation(resource.IamRole, role.IamRoleArn)
			}
			r.AddARNRelation(resource.KmsKey, cluster.KmsKeyId)

			if cluster.ElasticIpStatus != nil {
				r.AddRelation(resource.Ec2Eip, cluster.ElasticIpStatus.ElasticIp, "")
			}

			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
