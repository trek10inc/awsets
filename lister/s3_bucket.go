package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/fatih/structs"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

var listBucketsOnce sync.Once
var bucketsByRegion = make(map[string][]*types.Bucket)

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

	svc := s3.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	var outerErr error
	listBucketsOnce.Do(func() {

		res, err := svc.ListBuckets(ctx.Context, &s3.ListBucketsInput{})
		if err != nil {
			outerErr = fmt.Errorf("failed to query buckets from %s: %w", ctx.Region(), err)
			return
		}
		for i, bucket := range res.Buckets {
			bucketLocation, err := svc.GetBucketLocation(ctx.Context, &s3.GetBucketLocationInput{Bucket: bucket.Name})
			if err != nil {
				ctx.Logger.Errorf("failed to get bucket location for %s from %s: %w", *bucket.Name, ctx.Region(), err)
				continue
				//outerErr = fmt.Errorf("failed to get bucket location for %s from %s: %w", *bucket.Name, ctx.Region(), err)
				//return
			}
			reg := string(bucketLocation.LocationConstraint)
			if len(reg) == 0 {
				reg = "us-east-1"
			}
			bucketsByRegion[reg] = append(bucketsByRegion[reg], res.Buckets[i])
		}
	})
	if outerErr != nil {
		return rg, outerErr
	}

	for _, bucket := range bucketsByRegion[ctx.Region()] {
		//fmt.Printf("processing bucket %s in region %s\n", *bucket.Name, ctx.Region())
		buck := structs.Map(bucket)
		lifecycleRes, err := svc.GetBucketLifecycleConfiguration(ctx.Context, &s3.GetBucketLifecycleConfigurationInput{
			Bucket: bucket.Name,
		})
		if err == nil {
			buck["Lifecycle"] = structs.Map(lifecycleRes.Rules)
		} else {
			buck["Lifecycle"] = nil
		}

		//websiteRes, err := svc.GetBucketWebsite(ctx.Context, &s3.GetBucketWebsiteInput{
		//	Bucket: bucket.Name,
		//})

		policyRes, err := svc.GetBucketPolicy(ctx.Context, &s3.GetBucketPolicyInput{
			Bucket: bucket.Name,
		})
		if err == nil {
			buck["Policy"] = aws.ToString(policyRes.Policy)
		} else {
			buck["Policy"] = nil
		}

		encrRes, err := svc.GetBucketEncryption(ctx.Context, &s3.GetBucketEncryptionInput{
			Bucket: bucket.Name,
		})
		if err == nil {
			buck["Encryption"] = structs.Map(encrRes.ServerSideEncryptionConfiguration)
		} else {
			buck["Encryption"] = nil
		}

		tagRes, err := svc.GetBucketTagging(ctx.Context, &s3.GetBucketTaggingInput{
			Bucket: bucket.Name,
		})
		if err == nil {
			ts := structs.Map(tagRes.TagSet)
			buck["Tags"] = ts["TagSet"]
		} else {
			buck["Tags"] = nil
		}

		r := resource.New(ctx, resource.S3Bucket, bucket.Name, bucket.Name, buck)

		rg.AddResource(r)
	}
	return rg, nil
}
