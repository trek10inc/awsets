package awsets

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	context2 "github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/lister"
	"github.com/trek10inc/awsets/resource"
)

type ListerName string

type AWSets struct {
	AWSCfg    aws.Config
	AccountId string
	regions   []string
	Logger    context2.Logger
}

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
// returns a slice of the lister names that match. The filter is processed
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
	return nil, fmt.Errorf("no lister found for %s", name)
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
	return nil, fmt.Errorf("no lister found for %s", kind)
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
	regionsRes, err := ec2svc.DescribeRegions(context.Background(), &ec2.DescribeRegionsInput{
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
func List(ctx context2.AWSetsCtx, regions []string, listers []ListerName, cache Cacher) *resource.Group {

	if cache == nil {
		cache = NoopCache{}
	}

	// Creates a work queue
	jobs := make(chan job, 0)

	rg := resource.NewGroup()

	// Build worker pool
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int, workQueue <-chan job) {
			//var j job
			defer func() {
				ctx.Logger.Debugf("%d: finished worker\n", id)
				wg.Done()
			}()
			//defer func() {
			//	if r := recover(); r != nil {
			//		fmt.Printf("%d: Paniced in %s - %s. Error: %v\n", id, j.region, j.kind, r)
			//		fmt.Printf("%d: stacktrace from panic: %v\n", id, string(debug.Stack()))
			//	}
			//}()
			for {
				select {
				case job, more := <-workQueue:
					if !more {
						return
					}
					//j = job
					ctx.Logger.Debugf("%d: processing: %s - %s\n", id, job.region, job.lister)

					// If listing is cached, return it
					if cache.IsCached(job.region, job.lister) {
						ctx.Logger.Debugf("%d: cached: %s - %s\n", id, job.region, job.lister)
						group, err := cache.LoadGroup(job.region, job.lister)
						if err != nil {
							ctx.Logger.Errorf("failed to load group: %v", err)
						}
						rg.Merge(group)
					} else {
						ctx.Logger.Debugf("%d: not cached: %s - %s\n", id, job.region, job.lister)

						// Find the appropriate Lister
						l, err := GetByName(job.lister)
						if err != nil {
							ctx.Logger.Errorf("failed to get lister by name: %v", err)
							continue
						}

						// Copies the AWSets context - this also configures the region in the AWS config
						ctxcp := ctx.Copy(job.region)

						// Execute listing
						group, err := l.List(ctxcp)
						if err != nil {
							// indicates service is not supported in a region
							if strings.Contains(err.Error(), "no such host") {
								continue
							}
							ctx.Logger.Errorf("%d: failed job %s - %s\n with error: %v\n", id, job.region, job.lister, err)
							continue
						}

						// Update the results in the cache
						err = cache.SaveGroup(job.lister, group)
						if err != nil {
							ctx.Logger.Errorf("%d: failed to write cache for %s - %s: %v\n", id, job.region, job.lister, err)
						}

						// Merge the new results in with the rest
						rg.Merge(group)
					}
					ctx.Logger.Debugf("%d: complete: %s - %s\n", id, job.region, job.lister)
				}
			}
		}(i, jobs)
	}

	// Populate work queue with all Region/ListerName combinations
	for _, k := range listers {
		for _, r := range regions {
			jobs <- job{
				lister: k,
				region: r,
			}
		}
	}

	// Closes worker queue so the worker pool knows to stop
	close(jobs)

	wg.Wait()
	return rg
}

type job struct {
	lister ListerName
	region string
}

// Cacher is an interface that defines the necessary functions for an AWSets
// cache.
type Cacher interface {
	IsCached(region string, kind ListerName) bool
	SaveGroup(kind ListerName, group *resource.Group) error
	LoadGroup(region string, kind ListerName) (*resource.Group, error)
}

// NoopCache is the default cache provided by AWSets. It does nothing, and
// will never load nor save any data.
type NoopCache struct {
}

func (c NoopCache) IsCached(region string, kind ListerName) bool {
	return false
}

func (c NoopCache) SaveGroup(kind ListerName, group *resource.Group) error {
	return nil
}

func (c NoopCache) LoadGroup(region string, kind ListerName) (*resource.Group, error) {
	return resource.NewGroup(), nil
}
