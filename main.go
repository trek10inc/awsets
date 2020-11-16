package awsets

import (
	ctx2 "context"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/lister"
	"github.com/trek10inc/awsets/resource"
)

type ListerName string

// Types applies a filter to all supported AWS resources types and returns a
// slice of the ones that match. It first builds a list of all resources types
// that match any of the prefixes defined in `include`, and then removes any
// resource types that match any of the prefixes defined in `exclude`
func Types(include []string, exclude []string) []resource.ResourceType {
	filteredListers := Listers(include, exclude)

	typeMap := make(map[resource.ResourceType]struct{})
	for _, v := range filteredListers {
		for _, l := range lister.AllListers() {
			if v == ListerName(reflect.TypeOf(l).Name()) {
				for _, t := range l.Types() {
					typeMap[t] = struct{}{}
				}
			}
		}
	}

	ret := make([]resource.ResourceType, 0)
	for k := range typeMap {
		ret = append(ret, k)
	}
	return ret
}

// Listers applies an include/exclude filter to all implemented listers and
// returns a slice of the Lister names that match. The filter is processed
// against the resource types handled by each Lister.
func Listers(include []string, exclude []string) []ListerName {
	listerMap := make(map[ListerName]struct{}, 0)
	if len(include) == 0 {
		for _, v := range lister.AllListers() {
			listerMap[ListerName(reflect.TypeOf(v).Name())] = struct{}{}
		}
	} else {
		for _, name := range include {
			for _, v := range lister.AllListers() {
				for _, t := range v.Types() {
					if strings.HasPrefix(t.String(), name) {
						listerMap[ListerName(reflect.TypeOf(v).Name())] = struct{}{}
					}
				}
			}
		}
	}
	for _, name := range exclude {
		if len(name) == 0 {
			continue
		}
		for _, v := range lister.AllListers() {
			for _, t := range v.Types() {
				if strings.HasPrefix(t.String(), name) {
					delete(listerMap, ListerName(reflect.TypeOf(v).Name()))
				}
			}
		}
	}
	ret := make([]ListerName, 0)
	for v := range listerMap {
		ret = append(ret, v)
	}
	return ret
}

// GetByName finds the Lister that matches the name of the input argument. It
// returns an error if no match is found.
func GetByName(name ListerName) (lister.Lister, error) {

	for _, v := range lister.AllListers() {
		if name == ListerName(reflect.TypeOf(v).Name()) {
			return v, nil
		}
	}
	return nil, fmt.Errorf("no Lister found for %s", name)
}

// GetByType finds the Lister that processes the ResourceType given as an
// argument. It returns an error if no match is found.
func GetByType(kind resource.ResourceType) (lister.Lister, error) {

	for _, v := range lister.AllListers() {
		for _, t := range v.Types() {
			if t == kind {
				return v, nil
			}
		}
	}
	return nil, fmt.Errorf("no Lister found for %s", kind)
}

// Regions applies a filter to all available AWS regions and returns a list
// of the ones that match. The filtering is done by finding the regions that
// start with any of the prefixes pass in as arguments. If no prefixes are
// given, all available regions are returned.
func Regions(cfg aws.Config, prefixes ...string) ([]string, error) {

	// query AWS to find a list of all regions that the given credentials
	// have access to
	cfg.Region = "us-east-1"
	ec2svc := ec2.NewFromConfig(cfg)
	regionsRes, err := ec2svc.DescribeRegions(ctx2.Background(), &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query regions: %w", err)
	}

	// remove any AWS regions that are disabled in the current account
	regionMap := make(map[string]struct{}, 0)
	for _, r := range regionsRes.Regions {
		if r.OptInStatus != nil && *r.OptInStatus == "not-opted-in" {
			continue
		}
		if len(prefixes) == 0 {
			regionMap[*r.RegionName] = struct{}{}
		} else {
			for _, p := range prefixes {
				if strings.HasPrefix(*r.RegionName, p) {
					regionMap[*r.RegionName] = struct{}{}
				}
			}
		}
	}
	regions := make([]string, 0)
	for k := range regionMap {
		regions = append(regions, k)
	}
	return regions, nil
}

// List handles the execution of listers across multiple regions. It creates a
// worker pool to process every Lister/Region combination and aggregates the
// results together before returning them. If a cache is provided, each
// Lister/Region combination will first check for an existing result before
// querying AWS. Any new results will be updated in the cache.
func List(options ...Option) (*resource.Group, error) {

	// Create config struct, apply options
	awsetsCfg := &config{}
	for _, opt := range options {
		opt(awsetsCfg)
	}
	defer awsetsCfg.close()

	// Validates config values, defaults them if they are not specified
	err := awsetsCfg.validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	// Creates a work queue
	jobs := make(chan listjob, 0)
	totalJobs := len(awsetsCfg.Regions) * len(awsetsCfg.Listers)

	rg := resource.NewGroup()

	// Build worker pool
	wg := &sync.WaitGroup{}
	for i := 0; i < awsetsCfg.WorkerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case job, more := <-jobs:
					if !more {
						return
					}

					// Creates the AWSets context - this also configures the region in the AWS config
					workerCtx := createContext(awsetsCfg, id, job, totalJobs)

					workerCtx.SendStatus(context.StatusProcessing, "processing")
					group, err := processJob(workerCtx, id, job, awsetsCfg.Cache)
					if err != nil {
						workerCtx.SendStatus(context.StatusCompleteWithError, err.Error())
					} else {
						// Merge the new results in with the rest
						rg.Merge(group)
						workerCtx.SendStatus(context.StatusComplete, "")
					}
				}
			}
		}(i)
	}

	// Populate work queue with all Region/ListerName combinations
	for _, k := range awsetsCfg.Listers {
		for _, r := range awsetsCfg.Regions {
			jobs <- listjob{
				lister: k,
				region: r,
			}
		}
	}

	// Closes worker queue so the worker pool knows to stop
	close(jobs)

	wg.Wait()
	return rg, nil
}

// processJob does the actual querying of a particular resource type in a particular region. It catches panics caused by
// both the AWS SDK and also any other issues so that the full processing always completes.
func processJob(awsetsCtx *context.AWSetsCtx, id int, job listjob, cache Cacher) (rg *resource.Group, jobError error) {
	defer func() {
		if r := recover(); r != nil {
			jobError = fmt.Errorf("panicked: %v", r)
			awsetsCtx.SendStatus(context.StatusLogError, fmt.Sprintf("%d: stacktrace from panic: %v\n", id, string(debug.Stack())))
		}
	}()

	rg = resource.NewGroup()
	// If listing is cached, return it
	if cache.IsCached(job.region, job.lister) {
		awsetsCtx.SendStatus(context.StatusLogDebug, fmt.Sprintf("%d: cached: %s - %s", id, job.region, job.lister))
		rg, jobError = cache.LoadGroup(job.region, job.lister)
		if jobError != nil {
			jobError = fmt.Errorf("failed to load group from cache: %w", jobError)
		}
	} else {
		awsetsCtx.SendStatus(context.StatusLogDebug, fmt.Sprintf("%d: not cached: %s - %s", id, job.region, job.lister))

		// Find the appropriate Lister
		l, err := GetByName(job.lister)
		if err != nil {
			return rg, fmt.Errorf("failed to get lister by name %s: %v", job.lister, err)
		}

		// Execute listing
		rg, jobError = l.List(*awsetsCtx)
		if jobError != nil {
			// indicates service is not supported in a region
			if strings.Contains(jobError.Error(), "no such host") {
				jobError = nil
			}
			return
		}

		// Update the results in the cache
		jobError = cache.SaveGroup(job.lister, rg)
		if jobError != nil {
			jobError = fmt.Errorf("failed to write resources to cache: %w", jobError)
		}
	}
	return rg, jobError
}

type listjob struct {
	lister ListerName
	region string
}

func createContext(o *config, id int, job listjob, totalJobs int) *context.AWSetsCtx {
	ctx := &context.AWSetsCtx{
		AWSCfg:     o.AWSCfg.Copy(),
		AccountId:  o.AccountId,
		WorkerId:   id,
		Context:    o.Context,
		Lister:     string(job.lister),
		StatusChan: o.StatusChan,
		TotalJobs:  totalJobs,
	}
	ctx.AWSCfg.Region = job.region
	return ctx
}
