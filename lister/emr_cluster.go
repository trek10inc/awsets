package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"

	"github.com/aws/aws-sdk-go-v2/service/emr"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSEMRCluster struct {
}

func init() {
	i := AWSEMRCluster{}
	listers = append(listers, i)
}

func (l AWSEMRCluster) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.EmrCluster,
		resource.EmrInstanceFleetConfig,
		resource.EmrInstanceGroupConfig,
	}
}

func (l AWSEMRCluster) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := emr.New(ctx.AWSCfg)

	req := svc.ListClustersRequest(&emr.ListClustersInput{})
	rg := resource.NewGroup()
	paginator := emr.NewListClustersPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, id := range page.Clusters {
			cluster, err := svc.DescribeClusterRequest(&emr.DescribeClusterInput{
				ClusterId: id.Id,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to describe cluster %s: %w", *id.Id, err)
			}
			v := cluster.Cluster
			if v == nil {
				continue
			}
			r := resource.New(ctx, resource.EmrCluster, v.Id, v.Name, v)
			r.AddRelation(resource.EmrSecurityConfiguration, v.SecurityConfiguration, "")

			igPaginator := emr.NewListInstanceGroupsPaginator(svc.ListInstanceGroupsRequest(&emr.ListInstanceGroupsInput{
				ClusterId: id.Id,
			}))
			for igPaginator.Next(ctx.Context) {
				igPage := igPaginator.CurrentPage()
				for _, ig := range igPage.InstanceGroups {
					igR := resource.New(ctx, resource.EmrInstanceGroupConfig, ig.Id, ig.Name, ig)
					igR.AddRelation(resource.EmrCluster, v.Id, "")
					rg.AddResource(igR)
				}
			}
			if err = igPaginator.Err(); err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					if aerr.Code() == emr.ErrCodeInvalidRequestException &&
						strings.Contains(aerr.Message(), "Instance fleets and instance groups are mutually exclusive") {
						err = nil
					}
				}
				if err != nil {
					return rg, fmt.Errorf("failed to list instance groups for %s: %w", *v.Id, err)
				}
			}

			ifPaginator := emr.NewListInstanceFleetsPaginator(svc.ListInstanceFleetsRequest(&emr.ListInstanceFleetsInput{
				ClusterId: id.Id,
			}))
			for ifPaginator.Next(ctx.Context) {
				igPage := ifPaginator.CurrentPage()
				for _, fleet := range igPage.InstanceFleets {
					fleetR := resource.New(ctx, resource.EmrInstanceFleetConfig, fleet.Id, fleet.Name, fleet)
					fleetR.AddRelation(resource.EmrCluster, v.Id, "")

					for _, typeSpec := range fleet.InstanceTypeSpecifications {
						for _, bd := range typeSpec.EbsBlockDevices {
							fleetR.AddRelation(resource.Ec2Volume, bd.Device, "") // TODO: validate?
						}
					}

					rg.AddResource(fleetR)
				}
			}
			if err = ifPaginator.Err(); err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					if aerr.Code() == emr.ErrCodeInvalidRequestException &&
						strings.Contains(aerr.Message(), "Instance fleets and instance groups are mutually exclusive") {
						err = nil
					}
				}
				if err != nil {
					return rg, fmt.Errorf("failed to list instance fleets for %s: %w", *v.Id, err)
				}
			}

			stepsPaginator := emr.NewListStepsPaginator(svc.ListStepsRequest(&emr.ListStepsInput{
				ClusterId: id.Id,
			}))
			for stepsPaginator.Next(ctx.Context) {
				stepsPage := stepsPaginator.CurrentPage()
				for _, ss := range stepsPage.Steps {
					stepR := resource.New(ctx, resource.EmrStep, ss.Id, ss.Name, ss)
					stepR.AddRelation(resource.EmrCluster, v.Id, "")
					rg.AddResource(stepR)
				}
			}
			if err = stepsPaginator.Err(); err != nil {
				return rg, fmt.Errorf("failed to list steps for %s: %w", *v.Id, err)
			}

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
