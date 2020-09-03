package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appmesh"
	"github.com/trek10inc/awsets/context"
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
	}
}

func (l AWSAppMeshMesh) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := appmesh.New(ctx.AWSCfg)

	req := svc.ListMeshesRequest(&appmesh.ListMeshesInput{
		Limit: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := appmesh.NewListMeshesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, mesh := range page.Meshes {
			res, err := svc.DescribeMeshRequest(&appmesh.DescribeMeshInput{
				MeshName:  mesh.MeshName,
				MeshOwner: mesh.MeshOwner,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to describe mesh %s: %w", *mesh.MeshName, err)
			}
			v := res.Mesh
			if v == nil {
				continue
			}
			r := resource.New(ctx, resource.AppMeshMesh, v.MeshName, v.MeshName, v)

			// Virtual Routers
			vrPaginator := appmesh.NewListVirtualRoutersPaginator(svc.ListVirtualRoutersRequest(&appmesh.ListVirtualRoutersInput{
				Limit:     aws.Int64(100),
				MeshName:  mesh.MeshName,
				MeshOwner: mesh.MeshOwner,
			}))
			for vrPaginator.Next(ctx.Context) {
				vrPage := vrPaginator.CurrentPage()
				for _, vrId := range vrPage.VirtualRouters {
					vrRes, err := svc.DescribeVirtualRouterRequest(&appmesh.DescribeVirtualRouterInput{
						MeshName:          mesh.MeshName,
						MeshOwner:         mesh.MeshOwner,
						VirtualRouterName: vrId.VirtualRouterName,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to describe virtual router %s for mesh %s: %w", *vrId.VirtualRouterName, *mesh.MeshName, err)
					}
					if vr := vrRes.VirtualRouter; vr != nil {
						vrR := resource.New(ctx, resource.AppMeshVirtualRouter, vr.VirtualRouterName, vr.VirtualRouterName, vr)
						vrR.AddRelation(resource.AppMeshMesh, mesh.MeshName, "")

						routePaginator := appmesh.NewListRoutesPaginator(svc.ListRoutesRequest(&appmesh.ListRoutesInput{
							Limit:             aws.Int64(100),
							MeshName:          mesh.MeshName,
							MeshOwner:         mesh.MeshOwner,
							VirtualRouterName: vr.VirtualRouterName,
						}))
						for routePaginator.Next(ctx.Context) {
							routesPage := routePaginator.CurrentPage()
							for _, routeId := range routesPage.Routes {
								routeRes, err := svc.DescribeRouteRequest(&appmesh.DescribeRouteInput{
									MeshName:          mesh.MeshName,
									MeshOwner:         mesh.MeshOwner,
									RouteName:         routeId.RouteName,
									VirtualRouterName: vrId.VirtualRouterName,
								}).Send(ctx.Context)
								if err != nil {
									return rg, fmt.Errorf("failed to describe route %s for mesh %s: %w", *routeId.RouteName, *mesh.MeshName, err)
								}
								if route := routeRes.Route; route != nil {
									routeR := resource.New(ctx, resource.AppMeshRoute, route.RouteName, route.RouteName, route)
									routeR.AddRelation(resource.AppMeshMesh, mesh.MeshName, "")
									routeR.AddRelation(resource.AppMeshVirtualRouter, vr.VirtualRouterName, "")
									rg.AddResource(routeR)
								}
							}
						}
						if err = routePaginator.Err(); err != nil {
							return rg, fmt.Errorf("failed to list routes for virtual router %s and mesh %s: %w", *vr.VirtualRouterName, *mesh.MeshName, err)
						}

						rg.AddResource(vrR)
					}
				}
			}
			if err = vrPaginator.Err(); err != nil {
				return rg, fmt.Errorf("failed to list virtual routers for mesh %s: %w", *mesh.MeshName, err)
			}

			// Virtual Nodes
			vnPaginator := appmesh.NewListVirtualNodesPaginator(svc.ListVirtualNodesRequest(&appmesh.ListVirtualNodesInput{
				Limit:     aws.Int64(100),
				MeshName:  mesh.MeshName,
				MeshOwner: mesh.MeshOwner,
			}))
			for vnPaginator.Next(ctx.Context) {
				vnPage := vnPaginator.CurrentPage()
				for _, vnId := range vnPage.VirtualNodes {
					vnRes, err := svc.DescribeVirtualNodeRequest(&appmesh.DescribeVirtualNodeInput{
						MeshName:        mesh.MeshName,
						MeshOwner:       mesh.MeshOwner,
						VirtualNodeName: vnId.VirtualNodeName,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to describe virtual router %s for mesh %s: %w", *vnId.VirtualNodeName, *mesh.MeshName, err)
					}
					if vn := vnRes.VirtualNode; vn != nil {
						vnR := resource.New(ctx, resource.AppMeshVirtualNode, vn.VirtualNodeName, vn.VirtualNodeName, vn)
						vnR.AddRelation(resource.AppMeshMesh, mesh.MeshName, "")
						rg.AddResource(vnR)
					}
				}
			}
			if err = vnPaginator.Err(); err != nil {
				return rg, fmt.Errorf("failed to list virtual nodes for mesh %s: %w", *mesh.MeshName, err)
			}

			// Virtual Services
			vsPaginator := appmesh.NewListVirtualServicesPaginator(svc.ListVirtualServicesRequest(&appmesh.ListVirtualServicesInput{
				Limit:     aws.Int64(100),
				MeshName:  mesh.MeshName,
				MeshOwner: mesh.MeshOwner,
			}))
			for vsPaginator.Next(ctx.Context) {
				vsPage := vsPaginator.CurrentPage()
				for _, vsId := range vsPage.VirtualServices {
					vsRes, err := svc.DescribeVirtualServiceRequest(&appmesh.DescribeVirtualServiceInput{
						MeshName:           mesh.MeshName,
						MeshOwner:          mesh.MeshOwner,
						VirtualServiceName: vsId.VirtualServiceName,
					}).Send(ctx.Context)
					if err != nil {
						return rg, fmt.Errorf("failed to describe virtual service %s for mesh %s: %w", *vsId.VirtualServiceName, *mesh.MeshName, err)
					}
					if vs := vsRes.VirtualService; vs != nil {
						vsR := resource.New(ctx, resource.AppMeshVirtualService, vs.VirtualServiceName, vs.VirtualServiceName, vs)
						vsR.AddRelation(resource.AppMeshMesh, mesh.MeshName, "")
						rg.AddResource(vsR)
					}
				}
			}
			if err = vsPaginator.Err(); err != nil {
				return rg, fmt.Errorf("failed to list virtual services for mesh %s: %w", *mesh.MeshName, err)
			}

			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
