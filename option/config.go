package option

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AWSetsConfig struct {
	AWSCfg     aws.Config
	AccountId  string
	Context    context.Context
	Logger     Logger
	StatusChan chan<- StatusUpdate
}

func NewConfig(awsCfg aws.Config) (*AWSetsConfig, error) {
	awsCfg.Region = "us-east-1"
	svc := sts.NewFromConfig(awsCfg)
	res, err := svc.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return &AWSetsConfig{}, fmt.Errorf("failed to get account id: %w", err)
	}
	return &AWSetsConfig{
		AWSCfg:    awsCfg,
		AccountId: *res.Account,
		Context:   context.Background(),
		Logger:    DefaultLogger{},
	}, nil
}

func (c *AWSetsConfig) Region() string {
	return c.AWSCfg.Region
}

func (c *AWSetsConfig) Copy(region string) AWSetsConfig {
	cop := AWSetsConfig{
		AWSCfg:    c.AWSCfg.Copy(),
		AccountId: c.AccountId,
		Context:   c.Context,
		Logger:    c.Logger,
	}
	cop.AWSCfg.Region = region
	return cop
}

func (c *AWSetsConfig) SendStatus(update StatusUpdate) {
	if c.StatusChan == nil {
		return
	}
	c.StatusChan <- update
}

func (c *AWSetsConfig) Close() {
	if c.StatusChan != nil {
		close(c.StatusChan)
	}
}

type Option func(o *AWSetsConfig)

func WithLogger(logger Logger) Option {
	return func(o *AWSetsConfig) {
		o.Logger = logger
	}
}

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
