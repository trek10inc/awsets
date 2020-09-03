package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/trek10inc/awsets"
	"github.com/urfave/cli/v2"
)

var typesCmd = &cli.Command{
	Name:      "types",
	Usage:     "lists supported resource types",
	ArgsUsage: " ",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "include",
			Value: "",
			Usage: "comma separated list of resource type prefixes to include",
		},
		&cli.StringFlag{
			Name:  "exclude",
			Value: "",
			Usage: "comma separated list of resource type prefixes to exclude",
		},
	},
	Action: func(c *cli.Context) error {

		types := awspelunk.Types(strings.Split(c.String("include"), ","), strings.Split(c.String("exclude"), ","))
		ret := make([]string, 0)

		for _, t := range types {
			ret = append(ret, t.String())
		}

		sort.Strings(ret)

		for _, t := range ret {
			fmt.Printf("%s\n", t)
		}

		return nil
	},
}
