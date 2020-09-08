package lister

import (
	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/trek10inc/awsets/resource"
)

type AWSApiGatewayVpcLink struct {
}

func init() {
	i := AWSApiGatewayVpcLink{}
	listers = append(listers, i)
}

func (l AWSApiGatewayVpcLink) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ApiGatewayRestApi, resource.ApiGatewayDeployment, resource.ApiGatewayStage}
}

func (l AWSApiGatewayVpcLink) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := apigateway.New(ctx.AWSCfg)

	req := svc.GetVpcLinksRequest(&apigateway.GetVpcLinksInput{
		Limit: aws.Int64(500),
	})

	rg := resource.NewGroup()
	paginator := apigateway.NewGetVpcLinksPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Items {
			r := resource.New(ctx, resource.ApiGatewayVpcLink, v.Id, v.Name, v)
			rg.AddResource(r)
			// TODO: parse target ARNs to find relationships?
		}
	}
	err := paginator.Err()
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "AccessDeniedException" {
				// If api gateway is not supported in a region, returns access denied
				err = nil
			}
		}
	}
	return rg, err
}
