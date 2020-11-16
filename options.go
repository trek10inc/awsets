package awsets

import (
	ctx2 "context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/trek10inc/awsets/context"
)

// config is a struct that holds all the configuration values for the List method. This allows for a "functional option"
// approach to the API so it can be extended down the road without modifying the signature
type config struct {
	AWSCfg      *aws.Config
	AccountId   string
	WorkerCount int
	Context     ctx2.Context
	Regions     []string
	Listers     []ListerName
	Cache       Cacher
	StatusChan  chan<- context.StatusUpdate
}

func (c *config) close() {
	if c.StatusChan != nil {
		close(c.StatusChan)
	}
}

func (c *config) validate() error {

	if c.AWSCfg == nil {
		awsCfg, err := cfg.LoadDefaultConfig()
		if err != nil {
			return fmt.Errorf("failed to load AWS config: %w", err)
		}
		c.AWSCfg = &awsCfg
	}
	c.AWSCfg.Region = "us-east-1"
	svc := sts.NewFromConfig(*c.AWSCfg)
	res, err := svc.GetCallerIdentity(ctx2.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("failed to get account id: %w", err)
	}
	c.AccountId = *res.Account

	// Get Cache, default to NoOp if none is specified
	if c.Cache == nil {
		c.Cache = NoOpCache{}
	}
	// Initialize cache
	err = c.Cache.Initialize(c.AccountId)
	if err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}

	// Get regions, query all available if none are specified
	if len(c.Regions) == 0 {
		regions, err := Regions(*c.AWSCfg)
		if err != nil {
			return fmt.Errorf("failed to get regions: %w", err)
		}
		c.Regions = regions
	}

	// Get listers, default to all if none are specified
	if len(c.Listers) == 0 {
		c.Listers = Listers(nil, nil)
	}

	if c.Context == nil {
		c.Context = ctx2.Background()
	}

	if c.WorkerCount == 0 {
		c.WorkerCount = 10
	}
	return nil
}

type Option func(o *config)

// Creates new AWSets config struct with default values. It also queries AWS to get the current Account ID. Failures to
// query AWS can cause an error condition to be returned

func WithAWSConfig(awsCfg aws.Config) Option {
	return func(o *config) {
		o.AWSCfg = &awsCfg
	}
}

func WithContext(ctx ctx2.Context) Option {
	return func(o *config) {
		o.Context = ctx
	}
}

func WithRegions(regions []string) Option {
	return func(o *config) {
		o.Regions = regions
	}
}

func WithListers(listers []ListerName) Option {
	return func(o *config) {
		o.Listers = listers
	}
}

func WithCache(cache Cacher) Option {
	return func(o *config) {
		o.Cache = cache
	}
}

func WithStatus(ch chan<- context.StatusUpdate) Option {
	return func(o *config) {
		o.StatusChan = ch
	}
}

func WithWorkerCount(numWorkers int) Option {
	return func(o *config) {
		if numWorkers > 0 && numWorkers < 1000 {
			o.WorkerCount = numWorkers
		}
	}
}
