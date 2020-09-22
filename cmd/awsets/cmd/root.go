package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/urfave/cli/v2"
)

func Execute(buildInfo map[string]string) {
	app := &cli.App{
		Name:  "awsets",
		Usage: "query aws resources",
		Commands: []*cli.Command{
			listCmd,
			regionsCmd,
			typesCmd,
			processCmd,
			versionCmd(buildInfo),
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func configureAWS(ctx *cli.Context) (aws.Config, error) {
	if ctx.String("profile") != "" {
		return external.LoadDefaultAWSConfig(external.WithSharedConfigProfile(ctx.String("profile")))
	}
	return external.LoadDefaultAWSConfig()
}

func validateNumArgs(nArgs int) cli.BeforeFunc {
	return func(ctx *cli.Context) error {
		if ctx.NArg() != nArgs {
			return fmt.Errorf("expected %d arguments, but received %d", nArgs, ctx.NArg())
		}
		return nil
	}
}

func versionCmd(buildInfo map[string]string) *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "prints version information",
		Action: func(c *cli.Context) error {
			fmt.Printf("awsets - version: %s\tcommit: %s\tdate: %s\n", buildInfo["version"], buildInfo["commit"], buildInfo["date"])
			return nil
		},
	}
}
