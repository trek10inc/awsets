package lister

import (
	"fmt"
	"sync"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/fatih/structs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var listBucketsOnce sync.Once
var bucketsByRegion = make(map[string][]*s3.Bucket)

type AWSS3Bucket struct {
}

func init() {
	i := AWSS3Bucket{}
	listers = append(listers, i)
}

func (l AWSS3Bucket) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.S3Bucket}
}

func (l AWSS3Bucket) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := s3.New(ctx.AWSCfg)

	rg := resource.NewGroup()

	var outerErr error
	listBucketsOnce.Do(func() {

		req := svc.ListBucketsRequest(&s3.ListBucketsInput{})

		res, err := req.Send(ctx.Context)

		if err != nil {
			outerErr = fmt.Errorf("failed to query buckets from %s: %w", ctx.Region(), err)
			return
		}
		for i, bucket := range res.Buckets {
			bucketLocation, err := svc.GetBucketLocationRequest(&s3.GetBucketLocationInput{Bucket: bucket.Name}).Send(ctx.Context)
			if err != nil {
				ctx.Logger.Errorln(fmt.Errorf("failed to get bucket location for %s from %s: %w", *bucket.Name, ctx.Region(), err))
				continue
				//outerErr = fmt.Errorf("failed to get bucket location for %s from %s: %w", *bucket.Name, ctx.Region(), err)
				//return
			}
			reg, err := bucketLocation.LocationConstraint.MarshalValue()
			if err != nil {
				ctx.Logger.Errorln("failed to marshal s3 location for bucket %s: %v\n", *bucket.Name, err)
			}
			if len(reg) == 0 {
				reg = "us-east-1"
			}
			bucketsByRegion[reg] = append(bucketsByRegion[reg], &res.Buckets[i])
		}
	})
	if outerErr != nil {
		return rg, outerErr
	}

	for _, bucket := range bucketsByRegion[ctx.Region()] {
		//fmt.Printf("processing bucket %s in region %s\n", *bucket.Name, ctx.Region())
		buck := structs.Map(bucket)
		lifecycleRes, err := svc.GetBucketLifecycleConfigurationRequest(&s3.GetBucketLifecycleConfigurationInput{
			Bucket: bucket.Name,
		}).Send(ctx.Context)
		if err == nil {
			buck["Lifecycle"] = structs.Map(lifecycleRes.GetBucketLifecycleConfigurationOutput)
		} else {
			buck["Lifecycle"] = nil
		}

		//websiteRes, err := svc.GetBucketWebsiteRequest(&s3.GetBucketWebsiteInput{
		//	Bucket: bucket.Name,
		//}).Send(ctx.Context)

		policyRes, err := svc.GetBucketPolicyRequest(&s3.GetBucketPolicyInput{
			Bucket: bucket.Name,
		}).Send(ctx.Context)
		if err == nil {
			buck["Policy"] = aws.StringValue(policyRes.Policy)
		} else {
			buck["Policy"] = nil
		}

		encrRes, err := svc.GetBucketEncryptionRequest(&s3.GetBucketEncryptionInput{
			Bucket: bucket.Name,
		}).Send(ctx.Context)
		if err == nil {
			buck["Encryption"] = structs.Map(encrRes.GetBucketEncryptionOutput)
		} else {
			buck["Policy"] = nil
		}

		tagRes, err := svc.GetBucketTaggingRequest(&s3.GetBucketTaggingInput{
			Bucket: bucket.Name,
		}).Send(ctx.Context)
		if err == nil {
			ts := structs.Map(tagRes.GetBucketTaggingOutput)
			buck["Tags"] = ts["TagSet"]
		} else {
			buck["Tags"] = nil
		}

		r := resource.New(ctx, resource.S3Bucket, bucket.Name, bucket.Name, buck)

		rg.AddResource(r)
	}
	return rg, nil
}
