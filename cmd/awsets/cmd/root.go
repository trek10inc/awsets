package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func Execute(version string) {
	app := &cli.App{
		Name:  "awsets",
		Usage: "query aws resources",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Value: false,
				Usage: "enable verbose logging",
			},
		},
		Commands: []*cli.Command{
			listCmd,
			regionsCmd,
			typesCmd,
			processCmd,
			versionCmd(version),
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func versionCmd(version string) *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "prints version information",
		Action: func(c *cli.Context) error {
			fmt.Printf("awsets version: %s\n", version)
			return nil
		},
	}
}
