package awspelunk

import (
	"strings"
	"testing"
)

func Test_Types(t *testing.T) {
	types := Types([]string{""}, []string{""})
	if len(types) == 0 {
		t.Fatalf("expected all types")
	}
	types = Types(nil, nil)
	if len(types) == 0 {
		t.Fatalf("expected all types")
	}
	types = Types([]string{""}, []string{"ec2"})
	for _, rt := range types {
		if strings.HasPrefix(rt.String(), "ec2") {
			t.Fatalf("expected ec2* resource types to have been filtered out")
		}
	}
	types = Types([]string{"ec2"}, []string{""})
	for _, rt := range types {
		if !strings.HasPrefix(rt.String(), "ec2") {
			t.Fatalf("on expected ec2* resource types to be present")
		}
	}
}

func Test_Listers(t *testing.T) {
	listers := Listers([]string{""}, []string{""})
	if len(listers) == 0 {
		t.Fatalf("expected all listers")
	}
	listers = Listers(nil, nil)
	if len(listers) == 0 {
		t.Fatalf("expected all listers")
	}
}
