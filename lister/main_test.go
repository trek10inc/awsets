package lister

import (
	"testing"

	"github.com/trek10inc/awsets/resource"
)

func Test_DuplicateCheck(t *testing.T) {
	types := make(map[resource.ResourceType]struct{})
	dupes := make([]string, 0)
	for _, l := range listers {
		for _, kind := range l.Types() {
			if _, ok := types[kind]; ok {
				dupes = append(dupes, kind.String())
			}
			types[kind] = struct{}{}
		}
	}
	if len(dupes) > 0 {
		t.Fatalf("found duplicate types: %v\n", dupes)
	}
}
