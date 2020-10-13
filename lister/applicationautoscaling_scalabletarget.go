package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSApplicationAutoScalingScalableTarget struct {
}

func init() {
	i := AWSApplicationAutoScalingScalableTarget{}
	listers = append(listers, i)
}

func (l AWSApplicationAutoScalingScalableTarget) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ApplicationAutoScalingScalableTarget,
	}
}

func (l AWSApplicationAutoScalingScalableTarget) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := applicationautoscaling.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	for _, namespace := range []types.ServiceNamespace{
		types.ServiceNamespaceAppstream,
		types.ServiceNamespaceCassandra,
		types.ServiceNamespaceComprehend,
		types.ServiceNamespaceCustom_resource,
		types.ServiceNamespaceDynamodb,
		types.ServiceNamespaceLambda,
		types.ServiceNamespaceEc2,
		types.ServiceNamespaceEcs,
		types.ServiceNamespaceEmr,
		types.ServiceNamespaceRds,
		types.ServiceNamespaceSagemaker,
	} {
		err := Paginator(func(nt *string) (*string, error) {
			res, err := svc.DescribeScalableTargets(ctx.Context, &applicationautoscaling.DescribeScalableTargetsInput{
				MaxResults:       aws.Int32(50),
				ServiceNamespace: namespace,
				NextToken:        nt,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list scalable targets for %s: %w", namespace, err)
			}
			for _, v := range res.ScalableTargets {
				r := resource.New(ctx, resource.ApplicationAutoScalingScalableTarget, v.ResourceId, v.ResourceId, v)
				r.AddARNRelation(resource.IamRole, v.RoleARN)
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
