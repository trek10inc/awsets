package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
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

func (l AWSApiGatewayDomainName) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := apigateway.New(ctx.AWSCfg)

	req := svc.GetDomainNamesRequest(&apigateway.GetDomainNamesInput{
		Limit: aws.Int64(500),
	})

	rg := resource.NewGroup()
	paginator := apigateway.NewGetDomainNamesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, domainname := range page.Items {
			r := resource.New(ctx, resource.ApiGatewayDomainName, domainname.DomainName, domainname.DomainName, domainname)

			r.AddARNRelation(resource.AcmCertificate, domainname.CertificateArn)
			r.AddRelation(resource.Route53HostedZone, domainname.DistributionHostedZoneId, "")
			r.AddRelation(resource.Route53HostedZone, domainname.RegionalHostedZoneId, "")

			rg.AddResource(r)

			pathPaginor := apigateway.NewGetBasePathMappingsPaginator(svc.GetBasePathMappingsRequest(&apigateway.GetBasePathMappingsInput{
				DomainName: domainname.DomainName,
				Limit:      aws.Int64(500),
			}))
			for pathPaginor.Next(ctx.Context) {
				pathPage := pathPaginor.CurrentPage()
				for _, pathMapping := range pathPage.Items {
					pathRes := resource.New(ctx, resource.ApiGatewayBasePathMapping, pathMapping.BasePath, pathMapping.BasePath, pathMapping)
					pathRes.AddRelation(resource.ApiGatewayDomainName, domainname.DomainName, "")
					pathRes.AddRelation(resource.ApiGatewayStage, pathMapping.Stage, "")
					pathRes.AddRelation(resource.ApiGatewayRestApi, pathMapping.RestApiId, "")
					rg.AddResource(pathRes)
				}
			}
			if err := pathPaginor.Err(); err != nil {
				return rg, fmt.Errorf("failed to get base path mappings for %s: %w", aws.StringValue(domainname.DomainName), err)
			}
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
