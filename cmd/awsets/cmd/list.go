package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/trek10inc/awsets"
	"github.com/trek10inc/awsets/cmd/awsets/cache"
	context2 "github.com/trek10inc/awsets/context"
	"github.com/urfave/cli/v2"
)

var listCmd = &cli.Command{
	Name:      "list",
	Usage:     "lists all requested aws resources",
	ArgsUsage: " ",
	Before:    validateNumArgs(0),
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "dryrun",
			Value: false,
			Usage: "do a dry run of query",
		},
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
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Value:   "",
			Usage:   "output file to save results",
		},
		&cli.StringFlag{
			Name:  "profile",
			Value: "",
			Usage: "AWS profile to use",
		},
		&cli.BoolFlag{
			Name:  "refresh",
			Value: false,
			Usage: "force a refresh of cache",
		},
		&cli.StringFlag{
			Name:  "regions",
			Value: "",
			Usage: "comma separated list of region prefixes",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Value:   false,
			Usage:   "toggle verbose logging",
		},
	},
	Action: func(c *cli.Context) error {

		awscfg, err := configureAWS(c)
		if err != nil {
			log.Fatalf("failed to load aws config: %v\n", err)
		}
		var logger context2.Logger
		if c.Bool("verbose") {
			logger = VerboseLogger{}
		} else {
			logger = context2.DefaultLogger{}
		}

		ctx, err := context2.New(awscfg, context.Background(), logger)
		if err != nil {
			return fmt.Errorf("failed to initialize AWSets: %w", err)
		}

		listers := awsets.Listers(strings.Split(c.String("include"), ","), strings.Split(c.String("exclude"), ","))

		regions, err := awsets.Regions(awscfg, strings.Split(c.String("regions"), ",")...)
		if err != nil {
			return fmt.Errorf("failed to list regions: %w", err)
		}

		if c.Bool("dryrun") || c.Bool("verbose") {
			fmt.Printf("regions: %s\n", regions)
			types := awsets.Types(strings.Split(c.String("include"), ","), strings.Split(c.String("exclude"), ","))
			ret := make([]string, 0)
			for _, t := range types {
				ret = append(ret, t.String())
			}
			sort.Strings(ret)
			fmt.Printf("resource types: %v\n", ret)
		}
		if c.Bool("dryrun") {
			return nil
		}

		if c.Bool("verbose") {
			fmt.Printf("querying %d combinations\n", len(regions)*len(listers))
		}

		bc, err := cache.NewBoltCache(ctx.AccountId, c.Bool("refresh"))
		if err != nil {
			log.Fatalf("failed to open cache: %v", err)
		}

		rg := awsets.List(ctx, regions, listers, bc)

		j, err := rg.JSON()
		if err != nil {
			panic(err)
		}
		if c.String("output") == "" {
			fmt.Printf(j)
		} else {
			err = ioutil.WriteFile(c.String("output"), []byte(j), 0655)
			if err != nil {
				log.Fatalf("failed to write output file: %v\n", err)
			}
		}
		return nil
	},
}

type VerboseLogger struct {
}

func (l VerboseLogger) Errorf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func (l VerboseLogger) Infof(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, format, a...)
}

func (l VerboseLogger) Debugf(format string, a ...interface{}) {
	// No op for now until various levels of verbosity are supported
	//fmt.Fprintf(os.Stdout, format, a...)
}
