package main

import "github.com/trek10inc/awsets/cmd/awsets/cmd"

var (
	version string
)

func main() {
	cmd.Execute(version)
}
