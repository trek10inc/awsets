package option

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AWSetsConfig struct {
	AWSCfg      aws.Config
	WorkerCount int
	AccountId   string
	Context     context.Context
	StatusChan  chan<- StatusUpdate
}

func NewConfig(awsCfg aws.Config) (*AWSetsConfig, error) {
	awsCfg.Region = "us-east-1"
	svc := sts.NewFromConfig(awsCfg)
	res, err := svc.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return &AWSetsConfig{}, fmt.Errorf("failed to get account id: %w", err)
	}
	return &AWSetsConfig{
		AWSCfg:      awsCfg,
		AccountId:   *res.Account,
		Context:     context.Background(),
		WorkerCount: 10,
	}, nil
}

func (c *AWSetsConfig) Region() string {
	return c.AWSCfg.Region
}

func (c *AWSetsConfig) Copy(workerId, totalJobs int, region, lister string) *AWSetsConfig {
	ctx := c.Context
	ctx = context.WithValue(ctx, "workerId", workerId)
	ctx = context.WithValue(ctx, "totalJobs", totalJobs)
	ctx = context.WithValue(ctx, "region", region)
	ctx = context.WithValue(ctx, "lister", lister)

	cop := &AWSetsConfig{
		AWSCfg:     c.AWSCfg.Copy(),
		AccountId:  c.AccountId,
		Context:    ctx,
		StatusChan: c.StatusChan,
	}
	cop.AWSCfg.Region = region
	return cop
}

func (c *AWSetsConfig) CopyWithRegion(region string) *AWSetsConfig {
	ctx := c.Context
	ctx = context.WithValue(ctx, "region", region)

	cop := &AWSetsConfig{
		AWSCfg:     c.AWSCfg.Copy(),
		AccountId:  c.AccountId,
		Context:    ctx,
		StatusChan: c.StatusChan,
	}
	cop.AWSCfg.Region = region
	return cop
}

func (c *AWSetsConfig) SendStatus(statusType StatusType, msg string) {
	if c.StatusChan == nil {
		return
	}
	su := StatusUpdate{
		Type:      statusType,
		Lister:    c.Context.Value("lister").(string),
		Region:    c.Context.Value("region").(string),
		Message:   msg,
		WorkerId:  c.Context.Value("workerId").(int),
		TotalJobs: c.Context.Value("totalJobs").(int),
	}
	c.StatusChan <- su
}

func (c *AWSetsConfig) Close() {
	if c.StatusChan != nil {
		close(c.StatusChan)
	}
}

type Option func(o *AWSetsConfig)

func WithContext(ctx context.Context) Option {
	return func(o *AWSetsConfig) {
		o.Context = ctx
	}
}

func WithStatus(ch chan<- StatusUpdate) Option {
	return func(o *AWSetsConfig) {
		o.StatusChan = ch
	}
}

func WithWorkerCount(numWorkers int) Option {
	return func(o *AWSetsConfig) {
		if numWorkers > 0 && numWorkers < 1000 {
			o.WorkerCount = numWorkers
		}
	}
}
