package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSApplicationAutoScalingScalablePolicy struct {
}

func init() {
	i := AWSApplicationAutoScalingScalablePolicy{}
	listers = append(listers, i)
}

func (l AWSApplicationAutoScalingScalablePolicy) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ApplicationAutoScalingScalablePolicy,
	}
}

func (l AWSApplicationAutoScalingScalablePolicy) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := applicationautoscaling.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	for _, namespace := range []types.ServiceNamespace{
		types.ServiceNamespaceAppstream,
		types.ServiceNamespaceCassandra,
		types.ServiceNamespaceComprehend,
		types.ServiceNamespaceCustomResource,
		types.ServiceNamespaceDynamodb,
		types.ServiceNamespaceLambda,
		types.ServiceNamespaceEc2,
		types.ServiceNamespaceEcs,
		types.ServiceNamespaceEmr,
		types.ServiceNamespaceRds,
		types.ServiceNamespaceSagemaker,
	} {
		err := Paginator(func(nt *string) (*string, error) {
			res, err := svc.DescribeScalingPolicies(ctx.Context, &applicationautoscaling.DescribeScalingPoliciesInput{
				MaxResults:       aws.Int32(50),
				ServiceNamespace: namespace,
				NextToken:        nt,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list scalable policies for %s: %w", namespace, err)
			}
			for _, v := range res.ScalingPolicies {
				policyArn := arn.ParseP(v.PolicyARN)
				r := resource.New(ctx, resource.ApplicationAutoScalingScalablePolicy, policyArn.ResourceId, v.PolicyName, v)
				rg.AddResource(r)
			}
			return res.NextToken, nil
		})
		if err != nil {
			return rg, err
		}
	}

	return rg, nil
}
