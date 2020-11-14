package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSLambdaFunction struct {
}

func init() {
	i := AWSLambdaFunction{}
	listers = append(listers, i)
}

func (l AWSLambdaFunction) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.LambdaFunction, resource.LambdaVersion, resource.LambdaAlias, resource.LambdaEventSourceMapping}
}

func (l AWSLambdaFunction) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := lambda.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListFunctions(ctx.Context, &lambda.ListFunctionsInput{
			MaxItems: aws.Int32(100),
			Marker:   nt,
			// MasterRegion: TODO: maybe stringset this?
		})
		if err != nil {
			return nil, err
		}

		for _, function := range res.Functions {
			fxnArn := arn.ParseP(function.FunctionArn)
			r := resource.New(ctx, resource.LambdaFunction, fxnArn.ResourceId, function.Version, function)
			r.AddARNRelation(resource.KmsKey, function.KMSKeyArn)

			if function.VpcConfig != nil {
				if function.VpcConfig.VpcId != nil && *function.VpcConfig.VpcId != "" {
					r.AddRelation(resource.Ec2Vpc, function.VpcConfig.VpcId, "")
				}
				for _, sg := range function.VpcConfig.SecurityGroupIds {
					r.AddRelation(resource.Ec2SecurityGroup, sg, "")
				}
				for _, sn := range function.VpcConfig.SubnetIds {
					r.AddRelation(resource.Ec2Subnet, sn, "")
				}
			}
			for _, layer := range function.Layers {
				layerArn := arn.ParseP(layer.Arn)
				r.AddRelation(resource.LambdaLayerVersion, layerArn.ResourceId, layerArn.ResourceVersion)
			}

			// Event Source Mapping
			err = Paginator(func(nt2 *string) (*string, error) {
				sources, err := svc.ListEventSourceMappings(ctx.Context, &lambda.ListEventSourceMappingsInput{
					FunctionName: function.FunctionArn,
					MaxItems:     aws.Int32(100),
					Marker:       nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list event source mappings for %s: %w", *function.FunctionName, err)
				}
				for _, esm := range sources.EventSourceMappings {
					esmr := resource.New(ctx, resource.LambdaEventSourceMapping, esm.UUID, esm.UUID, esm)
					esmr.AddRelation(resource.LambdaFunction, fxnArn.ResourceId, function.Version)
					rg.AddResource(esmr)
					r.AddRelation(resource.LambdaEventSourceMapping, esm.UUID, "")
				}
				return sources.NextMarker, nil
			})
			if err != nil {
				return nil, err
			}

			// Invoke Configs

			eics := make([]*types.FunctionEventInvokeConfig, 0)
			err = Paginator(func(nt2 *string) (*string, error) {
				configs, err := svc.ListFunctionEventInvokeConfigs(ctx.Context, &lambda.ListFunctionEventInvokeConfigsInput{
					FunctionName: function.FunctionArn,
					MaxItems:     aws.Int32(50),
				})
				if err != nil {
					return nil, fmt.Errorf("failed to list event invoke configs for %s: %w", *function.FunctionName, err)
				}
				if len(configs.FunctionEventInvokeConfigs) > 0 {
					eics = append(eics, configs.FunctionEventInvokeConfigs...)
				}
				return configs.NextMarker, nil
			})
			if err != nil {
				return nil, err
			}
			if len(eics) > 0 {
				r.AddAttribute("EventInvokeConfigs", eics) // TODO: different way?
			}

			// Function Versions
			err = Paginator(func(nt2 *string) (*string, error) {
				versions, err := svc.ListVersionsByFunction(ctx.Context, &lambda.ListVersionsByFunctionInput{
					FunctionName: function.FunctionArn,
					MaxItems:     aws.Int32(100),
					Marker:       nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get versions for function %s: %w", *function.FunctionName, err)
				}

				for _, fv := range versions.Versions {
					fvArn := arn.ParseP(fv.FunctionArn)
					fvr := resource.New(ctx, resource.LambdaVersion, fvArn.ResourceId, fvArn.ResourceVersion, fv)
					fvr.AddRelation(resource.LambdaFunction, fxnArn.ResourceId, "")
					rg.AddResource(fvr)

					// Function Aliases
					err = Paginator(func(nt3 *string) (*string, error) {
						aliases, err := svc.ListAliases(ctx.Context, &lambda.ListAliasesInput{
							FunctionName:    fv.FunctionName,
							FunctionVersion: fv.Version,
							MaxItems:        aws.Int32(100),
							Marker:          nt3,
						})
						if err != nil {
							return nil, fmt.Errorf("failed to list function aliases for %s - %s: %w", *fv.FunctionName, *fv.Version, err)
						}
						for _, alias := range aliases.Aliases {
							aliasArn := arn.ParseP(alias.AliasArn)
							aliasRes := resource.New(ctx, resource.LambdaAlias, aliasArn.ResourceId, alias.Name, alias)
							aliasRes.AddRelation(resource.LambdaVersion, fvArn.ResourceId, fv.Version)
							rg.AddResource(aliasRes)
						}
						return aliases.NextMarker, nil
					})
					if err != nil {
						return nil, err
					}
				}
				return versions.NextMarker, nil
			})
			if err != nil {
				return nil, err
			}

			rg.AddResource(r)
		}
		return res.NextMarker, nil
	})
	return rg, err
}
