package command

import (
	"github.com/codegangsta/cli"
)

// Flags is a list of command-line options.
var Flags = []cli.Flag{
	cli.StringFlag{
		Name:  "e, environment",
		Usage: "Specify section/environment to parse from secrets.yaml.",
	},
	cli.StringFlag{
		Name:  "provider, p",
		Usage: "Name of the provider to use. May also be specified via SUMMON_PROVIDER environment variable.",
	},
	cli.StringFlag{
		Name:  "f",
		Value: "secrets.yml",
		Usage: "Path to secrets.yml.",
	},
	cli.StringSliceFlag{
		Name:  "D",
		Value: &cli.StringSlice{},
		Usage: "var=value causes substitution of value to $var.",
	},
	cli.StringFlag{
		Name:  "yaml",
		Usage: "secrets.yml as a literal string.",
	},
	cli.BoolFlag{
		Name:  "debug, d",
		Usage: "Print debug info during program execution. . May also be specified via SUMMON_DEBUG=true environment variable.",
	},
}
