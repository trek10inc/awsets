package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSApiGatewayV2DomainName struct {
}

func init() {
	i := AWSApiGatewayV2DomainName{}
	listers = append(listers, i)
}

func (l AWSApiGatewayV2DomainName) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ApiGatewayV2DomainName,
		resource.ApiGatewayV2ApiMapping,
	}
}

func (l AWSApiGatewayV2DomainName) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := apigatewayv2.New(ctx.AWSCfg)

	rg := resource.NewGroup()

	var nextToken *string
	for {
		res, err := svc.GetDomainNamesRequest(&apigatewayv2.GetDomainNamesInput{
			MaxResults: aws.String("100"),
			NextToken:  nextToken,
		}).Send(ctx.Context)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				if aerr.Code() == "AccessDeniedException" {
					// If api gateway is not supported in a region, returns access denied
					return rg, nil
				}
			}
			return rg, fmt.Errorf("failed to list apigatewayv2 domain names: %w", err)
		}
		for _, v := range res.Items {
			r := resource.New(ctx, resource.ApiGatewayV2DomainName, v.DomainName, v.DomainName, v)
			for _, dnc := range v.DomainNameConfigurations {
				r.AddRelation(resource.Route53HostedZone, dnc.HostedZoneId, "")
				r.AddARNRelation(resource.AcmCertificate, dnc.CertificateArn)
			}

			var mappingToken *string
			for {
				mappingRes, err := svc.GetApiMappingsRequest(&apigatewayv2.GetApiMappingsInput{
					DomainName: v.DomainName,
					MaxResults: aws.String("100"),
					NextToken:  mappingToken,
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list apigatewayv2 api mappings: %w", err)
				}
				for _, mapping := range mappingRes.Items {
					mappingR := resource.New(ctx, resource.ApiGatewayV2ApiMapping, mapping.ApiMappingId, mapping.ApiMappingId, mapping)
					mappingR.AddRelation(resource.ApiGatewayV2DomainName, v.DomainName, "")
					mappingR.AddRelation(resource.ApiGatewayV2Api, mapping.ApiId, "")
					rg.AddResource(mappingR)
				}
				if mappingRes.NextToken == nil {
					break
				}
				mappingToken = mappingRes.NextToken
			}

			rg.AddResource(r)
		}
		if res.NextToken == nil {
			break
		}
		nextToken = res.NextToken
	}
	return rg, nil
}
