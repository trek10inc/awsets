package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2VpcEndpoint struct {
}

func init() {
	i := AWSEc2VpcEndpoint{}
	listers = append(listers, i)
}

func (l AWSEc2VpcEndpoint) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.Ec2VpcEndpoint,
		resource.Ec2VpcEndpointConnectionNotification,
	}
}

func (l AWSEc2VpcEndpoint) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeVpcEndpoints(ctx.Context, &ec2.DescribeVpcEndpointsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.VpcEndpoints {
			r := resource.New(ctx, resource.Ec2VpcEndpoint, v.VpcEndpointId, v.VpcEndpointId, v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			for _, dns := range v.DnsEntries {
				r.AddRelation(resource.Route53HostedZone, dns.HostedZoneId, "")
			}
			for _, rt := range v.RouteTableIds {
				r.AddRelation(resource.Ec2RouteTable, rt, "")
			}
			for _, eni := range v.NetworkInterfaceIds {
				r.AddRelation(resource.Ec2NetworkInterface, eni, "")
			}
			for _, sg := range v.Groups {
				r.AddRelation(resource.Ec2SecurityGroup, sg.GroupId, "")
			}
			for _, sn := range v.SubnetIds {
				r.AddRelation(resource.Ec2Subnet, sn, "")
			}

			err = Paginator(func(nt2 *string) (*string, error) {
				ecns, err := svc.DescribeVpcEndpointConnectionNotifications(ctx.Context, &ec2.DescribeVpcEndpointConnectionNotificationsInput{
					ConnectionNotificationId: nil,
					DryRun:                   nil,
					Filters: []*types.Filter{{
						Name:   aws.String("vpc-endpoint-id"),
						Values: []*string{v.VpcEndpointId},
					}},
					MaxResults: aws.Int32(100),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get vpc endpoint connection notifications for %s: %w", *v.VpcEndpointId, err)
				}

				for _, ecn := range ecns.ConnectionNotificationSet {
					cnR := resource.New(ctx, resource.Ec2VpcEndpointConnectionNotification, ecn.ConnectionNotificationId, ecn.ConnectionNotificationId, ecn)
					cnR.AddRelation(resource.Ec2VpcEndpoint, ecn.VpcEndpointId, "")
					cnR.AddRelation(resource.Ec2VpcEndpointService, ecn.ServiceId, "")
					rg.AddResource(cnR)
				}

				return ecns.NextToken, nil
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
