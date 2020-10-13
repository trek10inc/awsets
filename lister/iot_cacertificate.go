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

	svc := iot.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	var marker *string
	for {
		cacerts, err := svc.ListCACertificates(ctx.Context, &iot.ListCACertificatesInput{
			PageSize: aws.Int32(100),
			Marker:   marker,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list iot ca certificates: %w", err)
		}
		for _, cacert := range cacerts.Certificates {
			r := resource.New(ctx, resource.IoTCACertificate, cacert.CertificateId, cacert.CertificateId, cacert)

			var certMarker *string
			for {
				certs, err := svc.ListCertificatesByCA(ctx.Context, &iot.ListCertificatesByCAInput{
					CaCertificateId: cacert.CertificateId,
					Marker:          certMarker,
					PageSize:        aws.Int32(100),
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list iot certificates for ca %s: %w", *cacert.CertificateId, err)
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
