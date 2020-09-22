package cmd

import (
	"fmt"
	"log"
	"sort"

	"github.com/trek10inc/awsets"
	"github.com/urfave/cli/v2"
)

var regionsCmd = &cli.Command{
	Name:      "regions",
	Usage:     "lists regions supported by account",
	ArgsUsage: "[region prefixes]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "profile",
			Value: "",
			Usage: "AWS profile to use",
		},
	},
	Action: func(c *cli.Context) error {

		awscfg, err := configureAWS(c)
		if err != nil {
			log.Fatalf("failed to load aws config: %v\n", err)
		}

		regions, err := awsets.Regions(awscfg, c.Args().Slice()...)
		if err != nil {
			log.Fatalf("failed to list regions: %v", err)
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
