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

func Paginator(f func(*string) (*string, error)) error {
	var nt *string
	for {
		t, err := f(nt)
		if err != nil {
			return err
		}
		if t == nil {
			break
		}
		nt = t
	}
	return nil
}
