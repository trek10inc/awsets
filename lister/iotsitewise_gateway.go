package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iotsitewise"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSIoTSiteWiseGateway struct {
}

func init() {
	i := AWSIoTSiteWiseGateway{}
	listers = append(listers, i)
}

func (l AWSIoTSiteWiseGateway) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.IoTSiteWiseGateway,
	}
}

func (l AWSIoTSiteWiseGateway) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := iotsitewise.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListGateways(ctx.Context, &iotsitewise.ListGatewaysInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, gateway := range res.GatewaySummaries {
			v, err := svc.DescribeGateway(ctx.Context, &iotsitewise.DescribeGatewayInput{
				GatewayId: gateway.GatewayId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe siteways gateway %s: %w", *gateway.GatewayId, err)
			}

			r := resource.New(ctx, resource.IoTSiteWiseGateway, v.GatewayId, v.GatewayName, v)
			if v.GatewayPlatform != nil {
				r.AddARNRelation(resource.GreengrassGroup, v.GatewayPlatform.Greengrass.GroupArn)
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
