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
	svc := databasemigrationservice.New(ctx.AWSCfg)

	req := svc.DescribeEndpointsRequest(&databasemigrationservice.DescribeEndpointsInput{
		MaxRecords: aws.Int64(100),
	})

	rg := resource.NewGroup()
	paginator := databasemigrationservice.NewDescribeEndpointsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, v := range page.Endpoints {
			r := resource.New(ctx, resource.DMSEndpoint, v.EndpointIdentifier, v.EndpointIdentifier, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			r.AddARNRelation(resource.AcmCertificate, v.CertificateArn)

			if setting := v.S3Settings; setting != nil {
				r.AddRelation(resource.S3Bucket, setting.BucketName, "")
			}
			r.AddARNRelation(resource.IamRole, v.ServiceAccessRoleArn)
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	if err != nil {
		if strings.Contains(err.Error(), "exceeded maximum number of attempts") {
			// If DMS is not supported in a region, it triggers this error
			err = nil
		}
	}
	return rg, err
}
