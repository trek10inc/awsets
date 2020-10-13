package lister

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSRedshiftSecurityGroup struct {
}

func init() {
	i := AWSRedshiftSecurityGroup{}
	listers = append(listers, i)
}

func (l AWSRedshiftSecurityGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.RedshiftSecurityGroup}
}

func (l AWSRedshiftSecurityGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := redshift.NewFromConfig(ctx.AWSCfg)

	res, err := svc.DescribeClusterSecurityGroups(ctx.Context, &redshift.DescribeClusterSecurityGroupsInput{
		MaxRecords: aws.Int32(100),
	})

	rg := resource.NewGroup()
	paginator := redshift.NewDescribeClusterSecurityGroupsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, sg := range page.ClusterSecurityGroups {
			r := resource.New(ctx, resource.RedshiftSecurityGroup, sg.ClusterSecurityGroupName, sg.ClusterSecurityGroupName, sg)
			for _, ec2sg := range sg.EC2SecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, ec2sg.EC2SecurityGroupName, "")
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	if strings.Contains(err.Error(), "VPC-by-Default customers cannot use cluster security groups") {
		// EC2-Classic thing
		err = nil
	}
	return rg, err
}
