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

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeClusterSecurityGroups(ctx.Context, &redshift.DescribeClusterSecurityGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "VPC-by-Default customers cannot use cluster security groups") {
				// EC2-Classic thing
				return nil, nil
			}
			return nil, err
		}
		for _, sg := range res.ClusterSecurityGroups {
			r := resource.New(ctx, resource.RedshiftSecurityGroup, sg.ClusterSecurityGroupName, sg.ClusterSecurityGroupName, sg)
			for _, ec2sg := range sg.EC2SecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, ec2sg.EC2SecurityGroupName, "")
			}
			rg.AddResource(r)
		}
		return res.Marker, nil
	})

	return rg, err
}
