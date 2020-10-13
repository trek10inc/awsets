package lister

import (
	"fmt"

	"github.com/trek10inc/awsets/context"

	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/trek10inc/awsets/arn"
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

	res, err := svc.ListFunctions(ctx.Context, &lambda.ListFunctionsInput{
		MaxItems: aws.Int32(100),
		// MasterRegion: TODO: maybe stringset this?
	})

	rg := resource.NewGroup()
	paginator := lambda.NewListFunctionsPaginator(req)
	for paginator.Next(ctx.Context) {
		page := paginator.CurrentPage()
		for _, function := range page.Functions {
			fxnArn := arn.ParseP(function.FunctionArn)
			r := resource.New(ctx, resource.LambdaFunction, fxnArn.ResourceId, function.Version, function)
			if function.KMSKeyArn != nil {
				keyArn := arn.ParseP(function.KMSKeyArn)
				r.AddRelation(resource.KmsKey, keyArn.ResourceId, "")
			}
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

			esmPaginator := lambda.NewListEventSourceMappingsPaginator(svc.ListEventSourceMappings(ctx.Context, &lambda.ListEventSourceMappingsInput{
				FunctionName: function.FunctionArn,
				MaxItems:     aws.Int32(100),
			}))
			for esmPaginator.Next(ctx.Context) {
				esmPage := esmPaginator.CurrentPage()
				for _, esm := range esmPage.EventSourceMappings {
					esmr := resource.New(ctx, resource.LambdaEventSourceMapping, esm.UUID, esm.UUID, esm)
					esmr.AddRelation(resource.LambdaFunction, fxnArn.ResourceId, function.Version)
					rg.AddResource(esmr)
					r.AddRelation(resource.LambdaEventSourceMapping, esm.UUID, "")
				}
			}
			if err := esmPaginator.Err(); err != nil {
				return nil, fmt.Errorf("failed to list event source mappings for %s: %w", *function.FunctionName, err)
			}

			eics := make([]lambda.FunctionEventInvokeConfig, 0)
			eicPaginator := lambda.NewListFunctionEventInvokeConfigsPaginator(svc.ListFunctionEventInvokeConfigsRequest(
				&lambda.ListFunctionEventInvokeConfigsInput{
					FunctionName: function.FunctionArn,
					MaxItems:     aws.Int32(50),
				},
			))
			for eicPaginator.Next(ctx.Context) {
				eicPage := eicPaginator.CurrentPage()
				if len(eicPage.FunctionEventInvokeConfigs) > 0 {
					eics = append(eics, eicPage.FunctionEventInvokeConfigs...)
				}
			}
			if err := eicPaginator.Err(); err != nil {
				return nil, fmt.Errorf("failed to list event invoke configs for %s: %w", *function.FunctionName, err)
			}
			if len(eics) > 0 {
				r.AddAttribute("EventInvokeConfigs", eics) // TODO: different way?
			}

			fvres, err := svc.ListVersionsByFunction(ctx.Context, &lambda.ListVersionsByFunctionInput{
				FunctionName: function.FunctionArn,
				MaxItems:     aws.Int32(100),
			})
			fvRes, err := fvReq
			if err != nil {
				return rg, err
			}
			for _, fv := range fvRes.Versions {
				fvArn := arn.ParseP(fv.FunctionArn)
				fvr := resource.New(ctx, resource.LambdaVersion, fvArn.ResourceId, fvArn.ResourceVersion, fv)
				fvr.AddRelation(resource.LambdaFunction, fxnArn.ResourceId, "")
				rg.AddResource(fvr)

				aliasPaginator := lambda.NewListAliasesPaginator(svc.ListAliases(ctx.Context, &lambda.ListAliasesInput{
					FunctionName:    fv.FunctionName,
					FunctionVersion: fv.Version,
					MaxItems:        aws.Int32(100),
				}))
				for aliasPaginator.Next(ctx.Context) {
					aliasPage := aliasPaginator.CurrentPage()
					for _, alias := range aliasPage.Aliases {
						aliasArn := arn.ParseP(alias.AliasArn)
						aliasRes := resource.New(ctx, resource.LambdaAlias, aliasArn.ResourceId, alias.Name, alias)
						aliasRes.AddRelation(resource.LambdaVersion, fvArn.ResourceId, fv.Version)
						rg.AddResource(aliasRes)
					}
				}
				err = aliasPaginator.Err()
				if err != nil {
					return nil, fmt.Errorf("failed to list function aliases for %s - %s: %w", *fv.FunctionName, *fv.Version, err)
				}
			}
			rg.AddResource(r)
		}
	}
	err := paginator.Err()
	return rg, err
}
