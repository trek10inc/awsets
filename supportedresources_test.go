package awsets

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/trek10inc/awsets/resource"
)

// Tests to make sure the documentation is up to date with the current list of supported resource types
func Test_SupportResources(t *testing.T) {
	// Read documented list of resource types
	f, err := os.Open("supported_resources.txt")
	if err != nil {
		t.Fatalf("failed to load file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	// Build map of types
	supported := make(map[resource.ResourceType]struct{})
	for scanner.Scan() {
		supported[resource.ResourceType(scanner.Text())] = struct{}{}
	}

	// Iterate through resource types supported in code. For each. check if it is in the documentation
	// If it is, remove it
	// If it isn't, append it to a list of resources that need added
	needsAdded := make([]resource.ResourceType, 0)
	for _, at := range Types(nil, nil) {
		_, exists := supported[at]
		if exists {
			delete(supported, at)
		} else {
			needsAdded = append(needsAdded, at)
		}
	}

	// If resources are missing from documentation, fail test & print them
	if len(needsAdded) > 0 {
		fmt.Printf("The following resource types need added to supported_types.txt:\n")
		for _, v := range needsAdded {
			fmt.Printf("%s\n", v)
		}
		t.Fail()
	}
	// If resources are in documentation that are NOT supported in code, fail test and print them
	if len(supported) > 0 {
		fmt.Printf("The following resource types need removed from supported_types.txt:\n")
		for _, v := range supported {
			fmt.Printf("%s\n", v)
		}
		t.Fail()
	}
}
