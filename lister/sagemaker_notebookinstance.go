package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/trek10inc/awsets/arn"
	"github.com/trek10inc/awsets/option"
	"github.com/trek10inc/awsets/resource"
)

type AWSSagemakerNotebookInstance struct {
}

func init() {
	i := AWSSagemakerNotebookInstance{}
	listers = append(listers, i)
}

func (l AWSSagemakerNotebookInstance) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.SagemakerNotebookInstance,
	}
}

func (l AWSSagemakerNotebookInstance) List(cfg option.AWSetsConfig) (*resource.Group, error) {
	svc := sagemaker.NewFromConfig(cfg.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListNotebookInstances(cfg.Context, &sagemaker.ListNotebookInstancesInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, ni := range res.NotebookInstances {
			v, err := svc.DescribeNotebookInstance(cfg.Context, &sagemaker.DescribeNotebookInstanceInput{
				NotebookInstanceName: ni.NotebookInstanceName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe sagemaker notebook instance %s: %w", *ni.NotebookInstanceName, err)
			}
			niArn := arn.ParseP(v.NotebookInstanceArn)
			r := resource.New(cfg, resource.SagemakerNotebookInstance, niArn.ResourceId, v.NotebookInstanceName, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			r.AddRelation(resource.Ec2Subnet, v.SubnetId, "")
			r.AddARNRelation(resource.IamRole, v.RoleArn)
			for _, sg := range v.SecurityGroups {
				r.AddRelation(resource.Ec2SecurityGroup, sg, "")
			}
			r.AddRelation(resource.Ec2NetworkInterface, v.NetworkInterfaceId, "")
			r.AddRelation(resource.SagemakerNotebookInstanceLifecycleConfig, v.NotebookInstanceLifecycleConfigName, "")
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
