package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
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

func (l AWSIoTCACertificate) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := iot.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListCACertificates(cfg.Context, &iot.ListCACertificatesInput{
			PageSize: aws.Int32(100),
			Marker:   nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list iot ca certificates: %w", err)
		}
		for _, cacert := range res.Certificates {
			r := resource.New(cfg, resource.IoTCACertificate, cacert.CertificateId, cacert.CertificateId, cacert)

			// Certs by CA
			err = Paginator(func(nt2 *string) (*string, error) {
				certs, err := svc.ListCertificatesByCA(cfg.Context, &iot.ListCertificatesByCAInput{
					CaCertificateId: cacert.CertificateId,
					Marker:          nt2,
					PageSize:        aws.Int32(100),
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list iot certificates for ca %s: %w", *cacert.CertificateId, err)
				}
				for _, cert := range certs.Certificates {
					r.AddRelation(resource.IoTCertificate, cert.CertificateId, "")
				}

				return res.NextMarker, nil
			})
			if err != nil {
				return nil, err
			}
			rg.AddResource(r)
		}
		return res.NextMarker, nil
	})
	return rg, err
}
