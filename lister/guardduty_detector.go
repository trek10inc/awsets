package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

type AWSGuardDutyDetector struct {
}

func init() {
	i := AWSGuardDutyDetector{}
	listers = append(listers, i)
}

func (l AWSGuardDutyDetector) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.GuardDutyDetector,
		resource.GuardDutyMember,
	}
}

func (l AWSGuardDutyDetector) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := guardduty.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListDetectors(ctx.Context, &guardduty.ListDetectorsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, id := range res.DetectorIds {
			v, err := svc.GetDetector(ctx.Context, &guardduty.GetDetectorInput{
				DetectorId: id,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get guard duty detector %s: %w", *id, err)
			}
			r := resource.New(ctx, resource.GuardDutyDetector, id, id, v)

			// Members
			err = Paginator(func(nt2 *string) (*string, error) {
				members, err := svc.ListMembers(ctx.Context, &guardduty.ListMembersInput{
					DetectorId:     id,
					MaxResults:     aws.Int32(100),
					NextToken:      nt2,
					OnlyAssociated: nil,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get members for guard duty detector %s: %w", *id, err)
				}

				for _, m := range members.Members {
					mR := resource.New(ctx, resource.GuardDutyMember, m.AccountId, m.AccountId, m)
					mR.AddRelation(resource.GuardDutyDetector, id, "")
					rg.AddResource(mR)
				}

				return members.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
