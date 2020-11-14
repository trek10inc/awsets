package awsets

import (
	ctx2 "context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/trek10inc/awsets/context"
)

// config is a struct that holds all the configuration values for the List method. This allows for a "functional option"
// approach to the API so it can be extended down the road without modifying the signature
type config struct {
	AWSCfg      aws.Config
	AccountId   string
	WorkerCount int
	Context     ctx2.Context
	Regions     []string
	Listers     []ListerName
	Cache       Cacher
	StatusChan  chan<- context.StatusUpdate
}

func (c *config) Close() {
	if c.StatusChan != nil {
		close(c.StatusChan)
	}
}

type Option func(o *config)

// Creates new AWSets config struct with default values. It also queries AWS to get the current Account ID. Failures to
// query AWS can cause an error condition to be returned
func newConfig(awsCfg aws.Config) (*config, error) {
	awsCfg.Region = "us-east-1"
	svc := sts.NewFromConfig(awsCfg)
	res, err := svc.GetCallerIdentity(ctx2.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return &config{}, fmt.Errorf("failed to get account id: %w", err)
	}
	return &config{
		AWSCfg:      awsCfg,
		AccountId:   *res.Account,
		Context:     ctx2.Background(),
		WorkerCount: 10,
	}, nil
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
