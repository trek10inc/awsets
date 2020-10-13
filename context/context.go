package context

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AWSetsCtx struct {
	AWSCfg    aws.Config
	AccountId string
	Context   context.Context
	Logger    Logger
}

func New(config aws.Config, ctx context.Context, logger Logger) (AWSetsCtx, error) {
	config.Region = "us-east-1"
	svc := sts.New(config)
	res, err := svc.GetCallerIdentity(ctx.Context, &sts.GetCallerIdentityInput{}).Send(context.Background())
	if err != nil {
		return AWSetsCtx{}, fmt.Errorf("failed to get account id: %w", err)
	}
	if logger == nil {
		logger = DefaultLogger{}
	}

	cfg := AWSetsCtx{
		AWSCfg:    config,
		AccountId: *res.Account,
		Context:   ctx,
		Logger:    logger,
	}

	return cfg, nil
}

func (c *AWSetsCtx) Region() string {
	return c.AWSCfg.Region
}

func (c *AWSetsCtx) Copy(region string) AWSetsCtx {
	cop := AWSetsCtx{
		AWSCfg:    c.AWSCfg.Copy(),
		AccountId: c.AccountId,
		Context:   c.Context,
		Logger:    c.Logger,
	}
	cop.AWSCfg.Region = region
	return cop
}
