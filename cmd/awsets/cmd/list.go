package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"

	context2 "github.com/trek10inc/awsets/context"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/trek10inc/awsets"
	"github.com/trek10inc/awsets/cmd/awsets/cache"
)

var listCmd = &cli.Command{
	Name:      "list",
	Usage:     "lists all requested aws resources",
	ArgsUsage: " ",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "dryrun",
			Value: false,
			Usage: "do a dry run of query",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Value:   "",
			Usage:   "output file to save results",
		},
		&cli.BoolFlag{
			Name:  "refresh",
			Value: false,
			Usage: "force a refresh of cache",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Value:   false,
			Usage:   "toggle verbose logging",
		},
		&cli.StringFlag{
			Name:  "regions",
			Value: "",
			Usage: "comma separated list of region prefixes",
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
	},
	Action: func(c *cli.Context) error {
		listers := awspelunk.Listers(strings.Split(c.String("include"), ","), strings.Split(c.String("exclude"), ","))

		regions, err := awspelunk.Regions(strings.Split(c.String("regions"), ",")...)
		if err != nil {
			log.Fatalf("unable to load regions: %v", err)
		}

		if c.Bool("dryrun") || c.Bool("verbose") {
			fmt.Printf("regions: %s\n", regions)
			types := awspelunk.Types(strings.Split(c.String("include"), ","), strings.Split(c.String("exclude"), ","))
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

		var logger context2.Logger
		if c.Bool("verbose") {
			logger = VerboseLogger{}
		} else {
			logger = context2.DefaultLogger{}
		}

		awscfg, err := external.LoadDefaultAWSConfig()
		if err != nil {
			log.Fatalf("failed to load aws config: %v\n", err)
		}
		ctx, err := context2.New(awscfg, context.Background(), logger)
		if err != nil {
			log.Fatalf("failed to create ctx: %v", err)
		}

		bc, err := cache.NewBoltCache(ctx.AccountId, c.Bool("refresh"))
		if err != nil {
			log.Fatalf("failed to open cache: %v", err)
		}

		rg, err := awspelunk.List(ctx, regions, listers, bc)
		if err != nil {
			log.Fatalf("failed to query resources? %v\n", err)
		}

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

func (l VerboseLogger) Errorln(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
}

func (l VerboseLogger) Infof(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, format, a...)
}

func (l VerboseLogger) Infoln(a ...interface{}) {
	fmt.Fprintln(os.Stdout, a...)
}
