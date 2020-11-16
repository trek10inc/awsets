package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSImageBuilderInfrastructureConfig struct {
}

func init() {
	i := AWSImageBuilderInfrastructureConfig{}
	listers = append(listers, i)
}

func (l AWSImageBuilderInfrastructureConfig) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ImageBuilderInfrastructureConfiguration,
	}
}

func (l AWSImageBuilderInfrastructureConfig) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := imagebuilder.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListInfrastructureConfigurations(ctx.Context, &imagebuilder.ListInfrastructureConfigurationsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list imagebuilder infrastructure configs: %w", err)
		}
		for _, config := range res.InfrastructureConfigurationSummaryList {
			configRes, err := svc.GetInfrastructureConfiguration(ctx.Context, &imagebuilder.GetInfrastructureConfigurationInput{
				InfrastructureConfigurationArn: config.Arn,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get imagebuilder infrastructure config %s: %w", *config.Name, err)
			}
			v := configRes.InfrastructureConfiguration
			configArn := arn.ParseP(v.Arn)
			r := resource.New(ctx, resource.ImageBuilderInfrastructureConfiguration, configArn.ResourceId, v.Name, v)
			for _, sg := range v.SecurityGroupIds {
				r.AddRelation(resource.Ec2SecurityGroup, sg, "")
			}
			r.AddRelation(resource.Ec2Subnet, v.SubnetId, "")
			r.AddARNRelation(resource.SnsTopic, v.SnsTopicArn)
			r.AddRelation(resource.Ec2KeyPair, v.KeyPair, "")
			r.AddARNRelation(resource.IamInstanceProfile, v.InstanceProfileName)
			if v.Logging != nil {
				if v.Logging.S3Logs != nil {
					r.AddRelation(resource.S3Bucket, v.Logging.S3Logs.S3BucketName, "")
				}
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
