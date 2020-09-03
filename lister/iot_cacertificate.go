package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/iot"
)

type AWSIoTCACertificate struct {
}

func init() {
	i := AWSIoTCACertificate{}
	listers = append(listers, i)
}

func (l AWSIoTCACertificate) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.IoTCACertificate}
}

func (l AWSIoTCACertificate) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := iot.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var marker *string
	for {
		cacerts, err := svc.ListCACertificatesRequest(&iot.ListCACertificatesInput{
			PageSize: aws.Int64(100),
			Marker:   marker,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list iot ca certificates: %w", err)
		}
		for _, cacert := range cacerts.Certificates {
			r := resource.New(ctx, resource.IoTCACertificate, cacert.CertificateId, cacert.CertificateId, cacert)

			var certMarker *string
			for {
				certs, err := svc.ListCertificatesByCARequest(&iot.ListCertificatesByCAInput{
					CaCertificateId: cacert.CertificateId,
					Marker:          certMarker,
					PageSize:        aws.Int64(100),
				}).Send(ctx.Context)
				if err != nil {
					return rg, fmt.Errorf("failed to list iot certificates for ca %s: %w", *cacert.CertificateId, err)
				}
				for _, cert := range certs.Certificates {
					r.AddRelation(resource.IoTCertificate, cert.CertificateId, "")
				}

				if certs.NextMarker == nil {
					break
				}
				certMarker = certs.NextMarker
			}

			rg.AddResource(r)
		}
		if cacerts.NextMarker == nil {
			break
		}
		marker = cacerts.NextMarker
	}
	return rg, nil
}
