package main

import (
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"go.temporal.io/aws-sdk-generator/internal"
)

func main() {
	err := run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	var templateDir, outputDir, service string

	app := cli.NewApp()
	app.Name = "temporal-aws-sdk-gen"
	app.Usage = "Generates Temporal Bindings for AWS SDK"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "template-dir",
			Required:    true,
			Usage:       "location of code generation template directory",
			Destination: &templateDir,
		},
		&cli.StringFlag{
			Name:        "output-dir",
			Required:    true,
			Usage:       "generated code location",
			Destination: &outputDir,
		},
		&cli.StringFlag{
			Name:        "service",
			Usage:       "service to regenerate, default is all services",
			Destination: &service,
		},
	}
	app.Action = func(c *cli.Context) (err error) {
		generator := internal.NewGenerator(templateDir)
		s := strings.ToLower(service)
		definitions, err := internal.ParseAwsSdk(s)
		if err != nil {
			return err
		}
		return generator.GenerateCode(outputDir, definitions)
	}
	return app.Run(args)
}
