package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSApiGatewayVpcLink struct {
}

func init() {
	i := AWSApiGatewayVpcLink{}
	listers = append(listers, i)
}

func (l AWSApiGatewayVpcLink) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ApiGatewayVpcLink}
}

func (l AWSApiGatewayVpcLink) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := apigateway.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.GetVpcLinks(cfg.Context, &apigateway.GetVpcLinksInput{
			Limit:    aws.Int32(500),
			Position: nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "AccessDeniedException") {
				// If api gateway is not supported in a region, returns access denied
				return nil, nil
			}
			return nil, fmt.Errorf("failed to get vpc links: %w", err)
		}
		for _, v := range res.Items {
			r := resource.New(cfg, resource.ApiGatewayVpcLink, v.Id, v.Name, v)
			rg.AddResource(r)
			// TODO: parse target ARNs to find relationships?
		}
		return res.Position, nil
	})
	return rg, err
}
