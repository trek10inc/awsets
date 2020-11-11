package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSSsmAssociation struct {
}

func init() {
	i := AWSSsmAssociation{}
	listers = append(listers, i)
}

func (l AWSSsmAssociation) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.SsmAssociation,
	}
}

func (l AWSSsmAssociation) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := ssm.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListAssociations(cfg.Context, &ssm.ListAssociationsInput{
			MaxResults: aws.Int32(50),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.Associations {
			r := resource.New(cfg, resource.SsmAssociation, v.AssociationId, v.AssociationName, v)
			r.AddRelation(resource.Ec2Instance, v.InstanceId, "")
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
