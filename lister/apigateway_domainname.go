package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSApiGatewayDomainName struct {
}

func init() {
	i := AWSApiGatewayDomainName{}
	listers = append(listers, i)
}

func (l AWSApiGatewayDomainName) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ApiGatewayDomainName, resource.ApiGatewayBasePathMapping}
}

func (l AWSApiGatewayDomainName) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := apigateway.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		req, err := svc.GetDomainNames(cfg.Context, &apigateway.GetDomainNamesInput{
			Limit:    aws.Int32(500),
			Position: nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "AccessDeniedException") {
				// If api gateway is not supported in a region, returns access denied
				return nil, nil
			}
			return nil, fmt.Errorf("failed to get domain names: %w", err)
		}
		for _, domainname := range req.Items {
			r := resource.New(cfg, resource.ApiGatewayDomainName, domainname.DomainName, domainname.DomainName, domainname)

			r.AddARNRelation(resource.AcmCertificate, domainname.CertificateArn)
			r.AddRelation(resource.Route53HostedZone, domainname.DistributionHostedZoneId, "")
			r.AddRelation(resource.Route53HostedZone, domainname.RegionalHostedZoneId, "")

			rg.AddResource(r)

			err = Paginator(func(nt2 *string) (*string, error) {
				pathRes, err := svc.GetBasePathMappings(cfg.Context, &apigateway.GetBasePathMappingsInput{
					DomainName: domainname.DomainName,
					Limit:      aws.Int32(500),
					Position:   nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get base path mappings for %s: %w", *domainname.DomainName, err)
				}
				for _, pathMapping := range pathRes.Items {
					pathRes := resource.New(cfg, resource.ApiGatewayBasePathMapping, pathMapping.BasePath, pathMapping.BasePath, pathMapping)
					pathRes.AddRelation(resource.ApiGatewayDomainName, domainname.DomainName, "")
					pathRes.AddRelation(resource.ApiGatewayStage, pathMapping.Stage, "")
					pathRes.AddRelation(resource.ApiGatewayRestApi, pathMapping.RestApiId, "")
					rg.AddResource(pathRes)
				}
				return pathRes.Position, nil
			})
		}
		return req.Position, nil
	})
	return rg, err
}
