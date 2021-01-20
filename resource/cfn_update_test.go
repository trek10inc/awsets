package resource

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"testing"
)

func Test_updatecfn(t *testing.T) {
	var cfSpec struct {
		PropertyTypes         map[string]interface{}
		ResourceTypes         map[string]interface{}
		ResourceSpecification string
	}

	// Read documented list of resource types
	res, err := http.Get("https://d1uauaxba7bl26.cloudfront.net/latest/gzip/CloudFormationResourceSpecification.json")
	if err != nil {
		t.Fatalf("failed to get spec: %v", err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&cfSpec)
	if err != nil {
		t.Fatalf("failed to decode spec: %v", err)
	}

	// Build map of types
	allCfn := make(map[string]struct{})
	for k := range cfSpec.ResourceTypes {
		allCfn[k] = struct{}{}
	}

	// Iterate through cloudformation types supported in code. For each. check if it is in current CF spec
	// If it is, remove it
	// If it isn't, append it to a list of resources that need added
	needsAdded := make([]string, 0)
	for cfn := range allCfn {
		_, exists := mapping[cfn]
		if exists {
			delete(mapping, cfn)
		} else {
			needsAdded = append(needsAdded, cfn)
		}
	}

	// If cloudformation resources are missing from code, fail test & print them
	if len(needsAdded) > 0 {
		fmt.Printf("The following CFN types need added:\n")
		sort.Strings(needsAdded)
		for _, v := range needsAdded {
			fmt.Printf("%s\n", v)
		}
		t.Fail()
	}
	// If cloudformation resources are in code that are NOT supported in the cloudformation spec, fail test and print them
	// Note: this should never be the case
	if len(mapping) > 0 {

		fmt.Printf("\n\nThe following CFN types need removed:\n")
		for k := range mapping {
			fmt.Printf("%s\n", k)
		}
		t.Fail()
	}
}
