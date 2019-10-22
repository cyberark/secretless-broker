package command

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/cyberark/summon/secretsyml"

	"github.com/cyberark/secretless-broker/internal/plugin"
	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/providers"
)

// The code in this file operates at the CLI level; it reads CLI arguments and will exit the process.

// Options contains the CLI arguments parsed by the cli framework.
type Options struct {
	Args        []string
	Filepath    string
	YamlInline  string
	Provider    string
	Subs        map[string]string
	Environment string
	Debug       bool
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
		return fmt.Errorf("Enter a subprocess to run")
	}

	commandArgs := &Options{
		Args:        c.Args(),
		Environment: c.String("environment"),
		Filepath:    c.String("f"),
		YamlInline:  c.String("yaml"),
		Provider:    c.String("provider"),
		Subs:        convertSubsToMap(c.StringSlice("D")),
		Debug:       c.Bool("debug"),
	}
	if !commandArgs.Debug {
		commandArgs.Debug = (os.Getenv("SUMMON_DEBUG") == "true")
	}

	if commandArgs.Provider == "" {
		commandArgs.Provider = os.Getenv("SUMMON_PROVIDER")
	}

	var err error
	var subcommand *Subcommand

	if subcommand, err = parseCommandArgsToSubcommand(commandArgs); err != nil {
		return err
	}

	subcommand.Stdout = c.App.Writer
	return subcommand.Run()
}

func parseCommandArgsToSubcommand(options *Options) (subcommand *Subcommand, err error) {
	subcommand = &Subcommand{Args: options.Args}

	if options.Provider == "" {
		err = fmt.Errorf("Provider option is required as a command argument (-provider or -p) or environment variable SUMMON_PROVIDER")
		return
	}

	// Load all internal Providers
	providerFactories := make(map[string]func(plugin_v1.ProviderOptions) (plugin_v1.Provider, error))
	for providerID, providerFactory := range providers.ProviderFactories {
		providerFactories[providerID] = providerFactory
	}

	resolver := plugin.NewResolver(providerFactories, nil, nil)

	if subcommand.Provider, err = resolver.Provider(options.Provider); err != nil {
		return
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
