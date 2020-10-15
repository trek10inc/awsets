package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
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

func (l AWSIoTCertificate) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := iot.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListCertificates(cfg.Context, &iot.ListCertificatesInput{
			PageSize: aws.Int32(100),
			Marker:   nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list iot certificates: %w", err)
		}
		for _, cert := range res.Certificates {
			r := resource.New(cfg, resource.IoTCertificate, cert.CertificateId, cert.CertificateId, cert)
			rg.AddResource(r)
		}
		return res.NextMarker, nil
	})
	return rg, err
}
