package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/iot"
)

type AWSIoTPolicy struct {
}

func init() {
	i := AWSIoTPolicy{}
	listers = append(listers, i)
}

func (l AWSIoTPolicy) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.IoTPolicy}
}

func (l AWSIoTPolicy) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := iot.New(ctx.AWSCfg)
	rg := resource.NewGroup()
	var marker *string
	for {
		policies, err := svc.ListPoliciesRequest(&iot.ListPoliciesInput{
			PageSize: aws.Int64(100),
			Marker:   marker,
		}).Send(ctx.Context)
		if err != nil {
			return rg, fmt.Errorf("failed to list iot policies: %w", err)
		}
		for _, policy := range policies.Policies {
			// TODO policy principals
			r := resource.New(ctx, resource.IoTPolicy, policy.PolicyName, policy.PolicyName, policy)
			rg.AddResource(r)
		}
		if policies.NextMarker == nil {
			break
		}
		marker = policies.NextMarker
	}
	return rg, nil
}
