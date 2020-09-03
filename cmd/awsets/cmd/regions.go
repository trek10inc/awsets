package cmd

import (
	"fmt"
	"sort"

	"github.com/trek10inc/awsets"
	"github.com/urfave/cli/v2"
)

var regionsCmd = &cli.Command{
	Name:      "regions",
	Usage:     "lists regions supported by account",
	ArgsUsage: "[region prefixes]",
	Action: func(c *cli.Context) error {

		regions, err := awspelunk.Regions(c.Args().Slice()...)

		if err != nil {
			return err
		}

		ret := make([]string, 0)

		for _, t := range regions {
			ret = append(ret, t)
		}

		sort.Strings(ret)

		for _, t := range ret {
			fmt.Printf("%s\n", t)
		}

		return nil
	},
}
