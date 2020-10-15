package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/emr"
	"github.com/trek10inc/awsets/option"
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

func (l AWSEMRCluster) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := emr.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListClusters(cfg.Context, &emr.ListClustersInput{
			Marker: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, id := range res.Clusters {
			cluster, err := svc.DescribeCluster(cfg.Context, &emr.DescribeClusterInput{
				ClusterId: id.Id,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe cluster %s: %w", *id.Id, err)
			}
			v := cluster.Cluster
			if v == nil {
				continue
			}
			r := resource.New(cfg, resource.EmrCluster, v.Id, v.Name, v)
			r.AddRelation(resource.EmrSecurityConfiguration, v.SecurityConfiguration, "")

			// Instance Groups
			err = Paginator(func(nt2 *string) (*string, error) {
				groups, err := svc.ListInstanceGroups(cfg.Context, &emr.ListInstanceGroupsInput{
					ClusterId: id.Id,
					Marker:    nt2,
				})
				if err != nil {
					if strings.Contains(err.Error(), "Instance fleets and instance groups are mutually exclusive") {
						return nil, nil
					}
					return nil, fmt.Errorf("failed to list instance groups for %s: %w", *v.Id, err)
				}
				for _, ig := range groups.InstanceGroups {
					igR := resource.New(cfg, resource.EmrInstanceGroupConfig, ig.Id, ig.Name, ig)
					igR.AddRelation(resource.EmrCluster, v.Id, "")
					rg.AddResource(igR)
				}
				return groups.Marker, nil
			})
			if err != nil {
				return nil, err
			}

			// Instance Fleets
			err = Paginator(func(nt2 *string) (*string, error) {
				fleets, err := svc.ListInstanceFleets(cfg.Context, &emr.ListInstanceFleetsInput{
					ClusterId: id.Id,
					Marker:    nt2,
				})
				if err != nil {
					if strings.Contains(err.Error(), "Instance fleets and instance groups are mutually exclusive") {
						return nil, nil
					}
					return nil, fmt.Errorf("failed to list instance fleets for %s: %w", *v.Id, err)
				}
				for _, fleet := range fleets.InstanceFleets {
					fleetR := resource.New(cfg, resource.EmrInstanceFleetConfig, fleet.Id, fleet.Name, fleet)
					fleetR.AddRelation(resource.EmrCluster, v.Id, "")

					for _, typeSpec := range fleet.InstanceTypeSpecifications {
						for _, bd := range typeSpec.EbsBlockDevices {
							fleetR.AddRelation(resource.Ec2Volume, bd.Device, "") // TODO: validate?
						}
					}

					rg.AddResource(fleetR)
				}
				return fleets.Marker, nil
			})
			if err != nil {
				return nil, err
			}

			// Steps
			err = Paginator(func(nt2 *string) (*string, error) {
				steps, err := svc.ListSteps(cfg.Context, &emr.ListStepsInput{
					ClusterId: id.Id,
					Marker:    nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list steps for %s: %w", *v.Id, err)
				}
				for _, ss := range steps.Steps {
					stepR := resource.New(cfg, resource.EmrStep, ss.Id, ss.Name, ss)
					stepR.AddRelation(resource.EmrCluster, v.Id, "")
					rg.AddResource(stepR)
				}
				return steps.Marker, nil
			})
			if err != nil {
				return nil, err
			}

			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
