package awsets

import "github.com/trek10inc/awsets/resource"

// Cacher is an interface that defines the necessary functions for an AWSets
// cache.
type Cacher interface {
	Initialize(accountId string) error
	IsCached(region string, kind ListerName) bool
	SaveGroup(kind ListerName, group *resource.Group) error
	LoadGroup(region string, kind ListerName) (*resource.Group, error)
}

// NoOpCache is the default cache provided by AWSets. It does nothing, and
// will never load nor save any data.
type NoOpCache struct {
}

func (c NoOpCache) Initialize(accountId string) error {
	return nil
}

func (c NoOpCache) IsCached(region string, kind ListerName) bool {
	return false
}

func (c NoOpCache) SaveGroup(kind ListerName, group *resource.Group) error {
	return nil
}

func (c NoOpCache) LoadGroup(region string, kind ListerName) (*resource.Group, error) {
	return resource.NewGroup(), nil
}
