package awspelunk

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	context2 "github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/trek10inc/awsets/lister"
	"github.com/trek10inc/awsets/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type ListerName string

func Types(include []string, exclude []string) []resource.ResourceType {
	typeMap := make(map[resource.ResourceType]struct{})
	if len(include) == 0 {
		for _, v := range lister.AllListers() {
			for _, t := range v.Types() {
				typeMap[t] = struct{}{}
			}
		}
	} else {
		for _, name := range include {

			for _, v := range lister.AllListers() {
				for _, t := range v.Types() {
					if strings.HasPrefix(t.String(), name) {
						typeMap[t] = struct{}{}
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
					delete(typeMap, t)
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

func Listers(include []string, exclude []string) []ListerName {
	idxMap := make(map[ListerName]struct{}, 0)
	if len(include) == 0 {
		for _, v := range lister.AllListers() {
			idxMap[ListerName(reflect.TypeOf(v).Name())] = struct{}{}
		}
	} else {
		for _, name := range include {
			for _, v := range lister.AllListers() {
				for _, t := range v.Types() {
					if strings.HasPrefix(t.String(), name) {
						idxMap[ListerName(reflect.TypeOf(v).Name())] = struct{}{}
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
					delete(idxMap, ListerName(reflect.TypeOf(v).Name()))
				}
			}
		}
	}
	ret := make([]ListerName, 0)
	for v := range idxMap {
		ret = append(ret, v)
	}
	return ret
}

func GetByName(name ListerName) lister.Lister {

	for _, v := range lister.AllListers() {
		if name == ListerName(reflect.TypeOf(v).Name()) {
			return v
		}
	}
	panic(fmt.Errorf("no lister found for %s", name))
}

func GetByType(kind resource.ResourceType) lister.Lister {

	for _, v := range lister.AllListers() {
		for _, t := range v.Types() {
			if t == kind {
				return v
			}
		}
	}
	panic(fmt.Errorf("no lister found for %s", kind))
}

func Regions(prefixes ...string) ([]string, error) {

	ctx, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}
	ctx.Region = "us-east-1"

	regionMap := make(map[string]struct{}, 0)
	ec2svc := ec2.New(ctx)
	regionsRes, err := ec2svc.DescribeRegionsRequest(&ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	}).Send(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to query regions: %w", err)
	}

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

func List(ctx context2.AWSetsCtx, regions []string, listers []ListerName, cache Cacher) (*resource.Group, error) {

	if cache == nil {
		cache = NoopCache{}
	}

	jobs := make(chan job, 0)

	rg := resource.NewGroup()
	wg := &sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int, workQueue <-chan job) {
			//var j job
			defer func() {
				ctx.Logger.Infof("%d: finished worker\n", id)
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
					ctx.Logger.Infof("%d: processing: %s - %s\n", id, job.region, job.lister)
					if cache.IsCached(ctx.AccountId, job.region, job.lister) {
						ctx.Logger.Infof("%d: cached: %s - %s\n", id, job.region, job.lister)
						group, err := cache.LoadGroup(job.region, job.lister)
						if err != nil {
							ctx.Logger.Errorf("failed to load group: %v", err)
						}
						rg.Merge(group)
					} else {
						ctx.Logger.Infof("%d: not cached: %s - %s\n", id, job.region, job.lister)
						idx := GetByName(job.lister)
						ctxcp := ctx.Copy(job.region)
						group, err := idx.List(ctxcp)
						if err != nil {
							// indicates service is not supported in a region, likely a better way to do this though
							// eks returns "AccessDenied" if the service isn't in the region though
							if strings.Contains(err.Error(), "no such host") {
								continue
							}
							ctx.Logger.Errorf("%d: failed job %s - %s\n with error: %v\n", id, job.region, job.lister, err)
							continue
						}
						err = cache.SaveGroup(group, job.region, job.lister)
						if err != nil {
							ctx.Logger.Errorf("%d: failed to write cache for %s - %s: %v\n", id, job.region, job.lister, err)
						}
						rg.Merge(group)
					}
					ctx.Logger.Infof("%d: complete: %s - %s\n", id, job.region, job.lister)
				}
			}
		}(i, jobs)
	}

	for _, k := range listers {
		for _, r := range regions {
			jobs <- job{
				lister: k,
				region: r,
			}
		}
	}
	close(jobs)

	wg.Wait()
	return rg, nil
}

type job struct {
	lister ListerName
	region string
}

type Cacher interface {
	IsCached(account, region string, kind ListerName) bool
	SaveGroup(group *resource.Group, region string, kind ListerName) error
	LoadGroup(region string, kind ListerName) (*resource.Group, error)
}

type NoopCache struct {
}

func (c NoopCache) IsCached(account, region string, kind ListerName) bool {
	return false
}

func (c NoopCache) SaveGroup(group *resource.Group, region string, kind ListerName) error {
	return nil
}

func (c NoopCache) LoadGroup(region string, kind ListerName) (*resource.Group, error) {
	return resource.NewGroup(), nil
}
