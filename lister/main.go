package lister

import (
	"github.com/trek10inc/awsets/context"
	"github.com/trek10inc/awsets/resource"
)

var listers = make([]Lister, 0)

type Lister interface {
	Types() []resource.ResourceType
	List(ctx context.AWSetsCtx) (*resource.Group, error)
}

func AllListers() []Lister {
	return listers
}
