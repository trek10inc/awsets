package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/trek10inc/awsets/option"
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

func (l AWSApiGatewayV2DomainName) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := apigatewayv2.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.GetDomainNames(cfg.Context, &apigatewayv2.GetDomainNamesInput{
			MaxResults: aws.String("100"),
			NextToken:  nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "AccessDeniedException") {
				// If api gateway is not supported in a region, returns access denied
				return nil, nil
			}
			return nil, fmt.Errorf("failed to list apigatewayv2 domain names: %w", err)
		}
		for _, v := range res.Items {
			r := resource.New(cfg, resource.ApiGatewayV2DomainName, v.DomainName, v.DomainName, v)
			for _, dnc := range v.DomainNameConfigurations {
				r.AddRelation(resource.Route53HostedZone, dnc.HostedZoneId, "")
				r.AddARNRelation(resource.AcmCertificate, dnc.CertificateArn)
			}

			// Mappings
			err = Paginator(func(nt2 *string) (*string, error) {
				mappingRes, err := svc.GetApiMappings(cfg.Context, &apigatewayv2.GetApiMappingsInput{
					DomainName: v.DomainName,
					MaxResults: aws.String("100"),
					NextToken:  nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list apigatewayv2 api mappings: %w", err)
				}
				for _, mapping := range mappingRes.Items {
					mappingR := resource.New(cfg, resource.ApiGatewayV2ApiMapping, mapping.ApiMappingId, mapping.ApiMappingId, mapping)
					mappingR.AddRelation(resource.ApiGatewayV2DomainName, v.DomainName, "")
					mappingR.AddRelation(resource.ApiGatewayV2Api, mapping.ApiId, "")
					rg.AddResource(mappingR)
				}
				return mappingRes.NextToken, nil
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
