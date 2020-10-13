package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/iot"
)

type AWSIoTCertificate struct {
}

func init() {
	i := AWSIoTCertificate{}
	listers = append(listers, i)
}

func (l AWSIoTCertificate) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.IoTCertificate}
}

func (l AWSIoTCertificate) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := iot.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	var marker *string
	for {
		certs, err := svc.ListCertificates(ctx.Context, &iot.ListCertificatesInput{
			PageSize: aws.Int32(100),
			Marker:   marker,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list iot certificates: %w", err)
		}
		for _, cert := range certs.Certificates {
			r := resource.New(ctx, resource.IoTCertificate, cert.CertificateId, cert.CertificateId, cert)
			rg.AddResource(r)
		}
		if certs.NextMarker == nil {
			break
		}
		marker = certs.NextMarker
	}
	return rg, nil
}
