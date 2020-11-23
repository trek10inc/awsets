package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSSignerSigningProfile struct {
}

func init() {
	i := AWSSignerSigningProfile{}
	listers = append(listers, i)
}

func (l AWSSignerSigningProfile) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.SignerSigningProfile}
}

func (l AWSSignerSigningProfile) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := signer.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListSigningProfiles(ctx.Context, &signer.ListSigningProfilesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.Profiles {
			profileArn := arn.ParseP(v.Arn)
			r := resource.New(ctx, resource.SignerSigningProfile, profileArn.ResourceId, v.ProfileName, v)

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
