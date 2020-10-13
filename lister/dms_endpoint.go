package lister

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSDMSEndpoint struct {
}

func init() {
	i := AWSDMSEndpoint{}
	listers = append(listers, i)
}

func (l AWSDMSEndpoint) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DMSEndpoint}
}

func (l AWSDMSEndpoint) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := databasemigrationservice.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeEndpoints(ctx.Context, &databasemigrationservice.DescribeEndpointsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "exceeded maximum number of attempts") {
				// If DMS is not supported in a region, it triggers this error
				return nil, nil
			}
			return nil, err
		}
		for _, v := range res.Endpoints {
			r := resource.New(ctx, resource.DMSEndpoint, v.EndpointIdentifier, v.EndpointIdentifier, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			r.AddARNRelation(resource.AcmCertificate, v.CertificateArn)

			if setting := v.S3Settings; setting != nil {
				r.AddRelation(resource.S3Bucket, setting.BucketName, "")
			}
			r.AddARNRelation(resource.IamRole, v.ServiceAccessRoleArn)
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
