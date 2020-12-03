package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSEc2TransitGateway struct {
}

func init() {
	i := AWSEc2TransitGateway{}
	listers = append(listers, i)
}

func (l AWSEc2TransitGateway) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.Ec2TransitGateway,
		resource.Ec2TransitGatewayAttachment,
	}
}

func (l AWSEc2TransitGateway) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeTransitGateways(ctx.Context, &ec2.DescribeTransitGatewaysInput{
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "is not valid for this web service") {
				// If transit gateways are not supported in a region, returns invalid action
				return nil, nil
			}
			return nil, err
		}
		for _, v := range res.TransitGateways {
			r := resource.New(ctx, resource.Ec2TransitGateway, v.TransitGatewayId, v.TransitGatewayId, v)
			// TODO lots of additional info to query here

			// Attachments
			err = Paginator(func(nt2 *string) (*string, error) {
				attachments, err := svc.DescribeTransitGatewayAttachments(ctx.Context, &ec2.DescribeTransitGatewayAttachmentsInput{
					Filters: []types.Filter{
						{
							Name:   aws.String("transit-gateway-id"),
							Values: []string{*v.TransitGatewayId},
						},
					},
					MaxResults: 100,
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get transit gateway attachments for %s: %w", *v.TransitGatewayId, err)
				}
				for _, a := range attachments.TransitGatewayAttachments {
					attachmentR := resource.New(ctx, resource.Ec2TransitGatewayAttachment, a.TransitGatewayAttachmentId, a.TransitGatewayAttachmentId, a)
					attachmentR.AddRelation(resource.Ec2TransitGateway, v.TransitGatewayId, "")
					switch a.ResourceType {
					case types.TransitGatewayAttachmentResourceTypeVpc:
						attachmentR.AddRelation(resource.Ec2Vpc, a.ResourceId, "")
					case types.TransitGatewayAttachmentResourceTypeVpn:
						attachmentR.AddRelation(resource.Ec2VpnGateway, a.ResourceId, "") // TODO validate type
					case types.TransitGatewayAttachmentResourceTypeDirectConnectGateway:
						//attachmentR.AddRelation(resource., a.ResourceId, "")
					case types.TransitGatewayAttachmentResourceTypePeering:
						attachmentR.AddRelation(resource.Ec2VpcPeering, a.ResourceId, "")
					case types.TransitGatewayAttachmentResourceTypeTgwPeering:
						//attachmentR.AddRelation(resource., a.ResourceId, "")
					}
					rg.AddResource(attachmentR)
				}
				return attachments.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Routes
			err = Paginator(func(nt2 *string) (*string, error) {
				routeTables, err := svc.DescribeTransitGatewayRouteTables(ctx.Context, &ec2.DescribeTransitGatewayRouteTablesInput{
					Filters: []types.Filter{
						{
							Name:   aws.String("transit-gateway-id"),
							Values: []string{*v.TransitGatewayId},
						},
					},
					MaxResults: 100,
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get transit gateway attachments for %s: %w", *v.TransitGatewayId, err)
				}

				for _, rt := range routeTables.TransitGatewayRouteTables {
					rtR := resource.New(ctx, resource.Ec2TransitGatewayRouteTable, rt.TransitGatewayRouteTableId, rt.TransitGatewayRouteTableId, rt)
					rtR.AddRelation(resource.Ec2TransitGateway, v.TransitGatewayId, "")
					// Route Table Associations
					allAssociations := make([]types.TransitGatewayRouteTableAssociation, 0)
					err = Paginator(func(nt3 *string) (*string, error) {
						associations, err := svc.GetTransitGatewayRouteTableAssociations(ctx.Context, &ec2.GetTransitGatewayRouteTableAssociationsInput{
							TransitGatewayRouteTableId: rt.TransitGatewayRouteTableId,
							MaxResults:                 100,
							NextToken:                  nt3,
						})
						if err != nil {
							return nil, fmt.Errorf("failed to get associations for route table %s: %w", *rt.TransitGatewayRouteTableId, err)
						}

						allAssociations = append(allAssociations, associations.Associations...)
						return associations.NextToken, nil
					})
					if err != nil {
						return nil, err
					}
					if len(allAssociations) > 0 {
						rtR.AddAttribute("Associations", allAssociations)
						for _, a := range allAssociations {
							switch a.ResourceType {
							case types.TransitGatewayAttachmentResourceTypeVpc:
								rtR.AddRelation(resource.Ec2Vpc, a.ResourceId, "")
							case types.TransitGatewayAttachmentResourceTypeVpn:
								rtR.AddRelation(resource.Ec2VpnGateway, a.ResourceId, "") // TODO validate type
							case types.TransitGatewayAttachmentResourceTypeDirectConnectGateway:
								//rtR.AddRelation(resource., a.ResourceId, "")
							case types.TransitGatewayAttachmentResourceTypePeering:
								rtR.AddRelation(resource.Ec2VpcPeering, a.ResourceId, "")
							case types.TransitGatewayAttachmentResourceTypeTgwPeering:
								//rtR.AddRelation(resource., a.ResourceId, "")
							}
						}
					}

					// Route Table Propagations
					allPropagations := make([]types.TransitGatewayRouteTablePropagation, 0)
					err = Paginator(func(nt3 *string) (*string, error) {
						propagations, err := svc.GetTransitGatewayRouteTablePropagations(ctx.Context, &ec2.GetTransitGatewayRouteTablePropagationsInput{
							TransitGatewayRouteTableId: rt.TransitGatewayRouteTableId,
							MaxResults:                 100,
							NextToken:                  nt3,
						})
						if err != nil {
							return nil, fmt.Errorf("failed to get propagations for route table %s: %w", *rt.TransitGatewayRouteTableId, err)
						}

						allPropagations = append(allPropagations, propagations.TransitGatewayRouteTablePropagations...)
						return propagations.NextToken, nil
					})
					if err != nil {
						return nil, err
					}
					if len(allPropagations) > 0 {
						rtR.AddAttribute("Propagations", allPropagations)
						for _, a := range allPropagations {
							switch a.ResourceType {
							case types.TransitGatewayAttachmentResourceTypeVpc:
								rtR.AddRelation(resource.Ec2Vpc, a.ResourceId, "")
							case types.TransitGatewayAttachmentResourceTypeVpn:
								rtR.AddRelation(resource.Ec2VpnGateway, a.ResourceId, "") // TODO validate type
							case types.TransitGatewayAttachmentResourceTypeDirectConnectGateway:
								//rtR.AddRelation(resource., a.ResourceId, "")
							case types.TransitGatewayAttachmentResourceTypePeering:
								rtR.AddRelation(resource.Ec2VpcPeering, a.ResourceId, "")
							case types.TransitGatewayAttachmentResourceTypeTgwPeering:
								//rtR.AddRelation(resource., a.ResourceId, "")
							}

						}
					}
					rg.AddResource(rtR)
				}
				return routeTables.NextToken, nil
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
