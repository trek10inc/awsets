package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appmesh"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSAppMeshMesh struct {
}

func init() {
	i := AWSAppMeshMesh{}
	listers = append(listers, i)
}

func (l AWSAppMeshMesh) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.AppMeshMesh,
		resource.AppMeshVirtualRouter,
		resource.AppMeshRoute,
		resource.AppMeshVirtualNode,
		resource.AppMeshVirtualService,
		resource.AppMeshVirtualGateway,
	}
}

func (l AWSAppMeshMesh) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := appmesh.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListMeshes(cfg.Context, &appmesh.ListMeshesInput{
			Limit:     aws.Int32(100),
			NextToken: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, mesh := range res.Meshes {
			res, err := svc.DescribeMesh(cfg.Context, &appmesh.DescribeMeshInput{
				MeshName:  mesh.MeshName,
				MeshOwner: mesh.MeshOwner,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe mesh %s: %w", *mesh.MeshName, err)
			}
			v := res.Mesh
			if v == nil {
				continue
			}
			r := resource.New(cfg, resource.AppMeshMesh, v.MeshName, v.MeshName, v)

			// Virtual Routers
			err = Paginator(func(nt2 *string) (*string, error) {
				routerRes, err := svc.ListVirtualRouters(cfg.Context, &appmesh.ListVirtualRoutersInput{
					Limit:     aws.Int32(100),
					MeshName:  mesh.MeshName,
					MeshOwner: mesh.MeshOwner,
					NextToken: nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list virtual routers for mesh %s: %w", *mesh.MeshName, err)
				}
				for _, vrId := range routerRes.VirtualRouters {
					vrRes, err := svc.DescribeVirtualRouter(cfg.Context, &appmesh.DescribeVirtualRouterInput{
						MeshName:          mesh.MeshName,
						MeshOwner:         mesh.MeshOwner,
						VirtualRouterName: vrId.VirtualRouterName,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to describe virtual router %s for mesh %s: %w", *vrId.VirtualRouterName, *mesh.MeshName, err)
					}
					if vr := vrRes.VirtualRouter; vr != nil {
						vrR := resource.New(cfg, resource.AppMeshVirtualRouter, vr.VirtualRouterName, vr.VirtualRouterName, vr)
						vrR.AddRelation(resource.AppMeshMesh, mesh.MeshName, "")

						err = Paginator(func(nt3 *string) (*string, error) {
							routesRes, err := svc.ListRoutes(cfg.Context, &appmesh.ListRoutesInput{
								Limit:             aws.Int32(100),
								MeshName:          mesh.MeshName,
								MeshOwner:         mesh.MeshOwner,
								VirtualRouterName: vr.VirtualRouterName,
							})
							if err != nil {
								return nil, fmt.Errorf("failed to list routes for virtual router %s and mesh %s: %w", *vr.VirtualRouterName, *mesh.MeshName, err)
							}
							for _, routeId := range routesRes.Routes {
								routeRes, err := svc.DescribeRoute(cfg.Context, &appmesh.DescribeRouteInput{
									MeshName:          mesh.MeshName,
									MeshOwner:         mesh.MeshOwner,
									RouteName:         routeId.RouteName,
									VirtualRouterName: vrId.VirtualRouterName,
								})
								if err != nil {
									return nil, fmt.Errorf("failed to describe route %s for mesh %s: %w", *routeId.RouteName, *mesh.MeshName, err)
								}
								if route := routeRes.Route; route != nil {
									routeR := resource.New(cfg, resource.AppMeshRoute, route.RouteName, route.RouteName, route)
									routeR.AddRelation(resource.AppMeshMesh, mesh.MeshName, "")
									routeR.AddRelation(resource.AppMeshVirtualRouter, vr.VirtualRouterName, "")
									rg.AddResource(routeR)
								}
							}
							return routesRes.NextToken, nil
						})
						rg.AddResource(vrR)
					}
				}

				return routerRes.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Virtual Nodes
			err = Paginator(func(nt2 *string) (*string, error) {
				nodesRes, err := svc.ListVirtualNodes(cfg.Context, &appmesh.ListVirtualNodesInput{
					Limit:     aws.Int32(100),
					MeshName:  mesh.MeshName,
					MeshOwner: mesh.MeshOwner,
					NextToken: nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list virtual nodes for mesh %s: %w", *mesh.MeshName, err)
				}
				for _, vnId := range nodesRes.VirtualNodes {
					vnRes, err := svc.DescribeVirtualNode(cfg.Context, &appmesh.DescribeVirtualNodeInput{
						MeshName:        mesh.MeshName,
						MeshOwner:       mesh.MeshOwner,
						VirtualNodeName: vnId.VirtualNodeName,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to describe virtual router %s for mesh %s: %w", *vnId.VirtualNodeName, *mesh.MeshName, err)
					}
					if vn := vnRes.VirtualNode; vn != nil {
						vnR := resource.New(cfg, resource.AppMeshVirtualNode, vn.VirtualNodeName, vn.VirtualNodeName, vn)
						vnR.AddRelation(resource.AppMeshMesh, mesh.MeshName, "")
						rg.AddResource(vnR)
					}
				}
				return nodesRes.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Virtual Services
			err = Paginator(func(nt2 *string) (*string, error) {
				servicesRes, err := svc.ListVirtualServices(cfg.Context, &appmesh.ListVirtualServicesInput{
					Limit:     aws.Int32(100),
					MeshName:  mesh.MeshName,
					MeshOwner: mesh.MeshOwner,
					NextToken: nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list virtual services for mesh %s: %w", *mesh.MeshName, err)
				}
				for _, vsId := range servicesRes.VirtualServices {
					vsRes, err := svc.DescribeVirtualService(cfg.Context, &appmesh.DescribeVirtualServiceInput{
						MeshName:           mesh.MeshName,
						MeshOwner:          mesh.MeshOwner,
						VirtualServiceName: vsId.VirtualServiceName,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to describe virtual service %s for mesh %s: %w", *vsId.VirtualServiceName, *mesh.MeshName, err)
					}
					if vs := vsRes.VirtualService; vs != nil {
						vsR := resource.New(cfg, resource.AppMeshVirtualService, vs.VirtualServiceName, vs.VirtualServiceName, vs)
						vsR.AddRelation(resource.AppMeshMesh, mesh.MeshName, "")
						rg.AddResource(vsR)
					}
				}
				return servicesRes.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Virtual Gateways
			err = Paginator(func(nt2 *string) (*string, error) {
				gwRes, err := svc.ListVirtualGateways(cfg.Context, &appmesh.ListVirtualGatewaysInput{
					Limit:     aws.Int32(100),
					MeshName:  mesh.MeshName,
					MeshOwner: mesh.MeshOwner,
					NextToken: nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list virtual gateways for mesh %s: %w", *mesh.MeshName, err)
				}
				for _, vgId := range gwRes.VirtualGateways {
					vgRes, err := svc.DescribeVirtualGateway(cfg.Context, &appmesh.DescribeVirtualGatewayInput{
						MeshName:           mesh.MeshName,
						MeshOwner:          mesh.MeshOwner,
						VirtualGatewayName: vgId.VirtualGatewayName,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to describe virtual gateways %s for mesh %s: %w", *vgId.VirtualGatewayName, *mesh.MeshName, err)
					}
					if vg := vgRes.VirtualGateway; vg != nil {
						vgR := resource.New(cfg, resource.AppMeshVirtualGateway, vg.VirtualGatewayName, vg.VirtualGatewayName, vg)
						vgR.AddRelation(resource.AppMeshMesh, mesh.MeshName, "")
						rg.AddResource(vgR)
					}
				}

				return gwRes.NextToken, nil
			})

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
