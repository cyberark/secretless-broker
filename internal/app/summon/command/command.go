package command

import (
	"fmt"
	"io"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/conjurinc/secretless/internal/app/secretless"
	"github.com/conjurinc/secretless/internal/pkg/provider"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	"github.com/cyberark/summon/secretsyml"
)

// The code in this file operates at the CLI level; it reads CLI arguments and will exit the process.

// Options contains the CLI arguments parsed by the cli framework.
type Options struct {
	Args        []string
	Filepath    string
	YamlInline  string
	Subs        map[string]string
	ConfigFile  string
	Environment string
}

// VERSION is the semantic version.
const VERSION = "0.1.0"

// RunCLI defines and runs the command line program.
func RunCLI(args []string, writer io.Writer) error {
	app := cli.NewApp()
	app.Name = "summon2"
	app.Usage = "Parse secrets.yml and export environment variables"
	app.Version = VERSION
	app.Writer = writer
	app.Flags = Flags
	app.Action = Action

	return app.Run(args)
}

// Action is the main entry point for the CLI command.
var Action = func(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("Enter a subprocess to run!")
	}

	commandArgs := &Options{
		Args:        c.Args(),
		Environment: c.String("environment"),
		Filepath:    c.String("f"),
		YamlInline:  c.String("yaml"),
		ConfigFile:  c.String("config"),
		Subs:        convertSubsToMap(c.StringSlice("D")),
	}

	var err error
	var subcommand *Subcommand

	if subcommand, err = parseCommandArgsToSubcommand(commandArgs); err != nil {
		return err
	}

	return subcommand.Run()
}

func parseCommandArgsToSubcommand(options *Options) (subcommand *Subcommand, err error) {
	subcommand = &Subcommand{Args: options.Args}

	if options.ConfigFile != "" {
		config := config.Configure(options.ConfigFile)
		providers := make([]provider.Provider, len(config.Providers))
		for i := range config.Providers {
			if providers[i], err = secretless.LoadProvider(config.Providers[i]); err != nil {
				err = fmt.Errorf("Unable to load provider '%s' : %s", config.Providers[i].Name, err.Error())
				return
			}
		}
		subcommand.Providers = providers
	}

	var secrets secretsyml.SecretsMap

	switch options.YamlInline {
	case "":
		secrets, err = secretsyml.ParseFromFile(options.Filepath, options.Environment, options.Subs)
	default:
		secrets, err = secretsyml.ParseFromString(options.YamlInline, options.Environment, options.Subs)
	}

	if err != nil {
		return
	}

	subcommand.SecretsMap = secrets

	return
}

// convertSubsToMap converts the list of substitutions passed in via
// command line to a map
func convertSubsToMap(subs []string) map[string]string {
	out := make(map[string]string)
	for _, sub := range subs {
		s := strings.SplitN(sub, "=", 2)
		key, val := s[0], s[1]
		out[key] = val
	}
	return out
}
