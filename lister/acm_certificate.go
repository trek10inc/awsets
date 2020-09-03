package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/resource"
)

type AWSAcmCertificate struct {
}

func init() {
	i := AWSAcmCertificate{}
	listers = append(listers, i)
}

func (l AWSAcmCertificate) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.AcmCertificate}
}

func (l AWSAcmCertificate) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := acm.New(ctx.AWSCfg)

	req := svc.ListCertificatesRequest(&acm.ListCertificatesInput{
		MaxItems: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := acm.NewListCertificatesPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, cert := range page.CertificateSummaryList {
			res, err := svc.DescribeCertificateRequest(&acm.DescribeCertificateInput{CertificateArn: cert.CertificateArn}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("unable to describe certificate %s: %w", *cert.CertificateArn, err)
			}
			//if arn.IsArnP(res.Certificate.CertificateArn) {
			certArn := arn.ParseP(res.Certificate.CertificateArn)
			r := resource.New(ctx, resource.AcmCertificate, certArn.ResourceId, certArn.ResourceId, res.Certificate)
			//}
			tagRes, err := svc.ListTagsForCertificateRequest(&acm.ListTagsForCertificateInput{
				CertificateArn: cert.CertificateArn,
			}).Send(ctx.Context)
			if err != nil {
				return rg, fmt.Errorf("failed to list tags for cert %s: %w", *cert.CertificateArn, err)
			}
			for _, tag := range tagRes.Tags {
				r.Tags[*tag.Key] = *tag.Value
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
