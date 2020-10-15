package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
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

func (l AWSIoTPolicy) List(cfg option.AWSetsConfig) (*resource.Group, error) {

	svc := iot.NewFromConfig(cfg.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListPolicies(cfg.Context, &iot.ListPoliciesInput{
			PageSize: aws.Int32(100),
			Marker:   nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list iot policies: %w", err)
		}
		for _, policy := range res.Policies {
			// TODO policy principals
			r := resource.New(cfg, resource.IoTPolicy, policy.PolicyName, policy.PolicyName, policy)
			rg.AddResource(r)
		}
		return res.NextMarker, nil
	})
	return rg, err
}
