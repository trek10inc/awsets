package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSTransferServer struct {
}

func init() {
	i := AWSTransferServer{}
	listers = append(listers, i)
}

func (l AWSTransferServer) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.TransferServer,
		resource.TransferUser,
	}
}

func (l AWSTransferServer) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := transfer.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListServers(ctx.Context, &transfer.ListServersInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, server := range res.Servers {
			v, err := svc.DescribeServer(ctx.Context, &transfer.DescribeServerInput{
				ServerId: server.ServerId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get transfer server %s: %w", *server.ServerId, err)
			}
			r := resource.New(ctx, resource.TransferServer, v.Server.ServerId, v.Server.ServerId, v.Server)
			r.AddARNRelation(resource.IamRole, v.Server.LoggingRole)
			if ed := v.Server.EndpointDetails; ed != nil {
				r.AddRelation(resource.Ec2Vpc, ed.VpcId, "")
				for _, sn := range ed.SubnetIds {
					r.AddRelation(resource.Ec2Subnet, sn, "")
				}
			}

			// Transfer Users
			err = Paginator(func(nt2 *string) (*string, error) {
				users, err := svc.ListUsers(ctx.Context, &transfer.ListUsersInput{
					ServerId:   v.Server.ServerId,
					MaxResults: aws.Int32(100),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list transfer users for server %s: %w", *v.Server.ServerId, err)
				}
				for _, listeduser := range users.Users {
					ud, err := svc.DescribeUser(ctx.Context, &transfer.DescribeUserInput{
						ServerId: v.Server.ServerId,
						UserName: listeduser.UserName,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to describe transfer user %s for server %s: %w", *listeduser.UserName, *v.Server.ServerId, err)
					}
					uRes := resource.New(ctx, resource.TransferUser, ud.User.UserName, ud.User.UserName, ud.User)
					uRes.AddRelation(resource.TransferServer, v.Server.ServerId, "")
					rg.AddResource(uRes)
				}

				return users.NextToken, nil
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
