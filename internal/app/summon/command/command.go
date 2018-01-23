package command

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/cyberark/summon/secretsyml"
	"github.com/kgilpin/secretless/internal/app/secretless"
	"github.com/kgilpin/secretless/internal/pkg/provider"
	"github.com/kgilpin/secretless/pkg/secretless/config"
	yaml "gopkg.in/yaml.v1"
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
var Action = func(c *cli.Context) {
	if !c.Args().Present() {
		fmt.Println("Enter a subprocess to run!")
		os.Exit(127)
	}

	commandArgs := &Options{
		Args:        c.Args(),
		Environment: c.String("environment"),
		Filepath:    c.String("f"),
		YamlInline:  c.String("yaml"),
		ConfigFile:  c.String("config"),
		Subs:        convertSubsToMap(c.StringSlice("D")),
		Debug:       c.Bool("debug"),
	}

	var err error
	var subcommand *Subcommand
	var out string

	if subcommand, err = parseCommandArgsToSubcommand(commandArgs); err != nil {
		fmt.Println(err.Error())
		os.Exit(127)
	}

	out, err = subcommand.Run()

	code, err := returnStatusOfError(err)

	if err != nil {
		fmt.Println(out + ": " + err.Error())
		os.Exit(127)
	}

	os.Exit(code)
}

func parseCommandArgsToSubcommand(options *Options) (subcommand *Subcommand, err error) {
	subcommand = &Subcommand{Args: options.Args}

	var c config.Config

	if options.ConfigFile != "" {
		c = config.Configure(options.ConfigFile)
	}

	if options.Debug {
		configStr, _ := yaml.Marshal(c)
		log.Printf("Loaded configuration : %s", configStr)
	}

	providers := make([]provider.Provider, len(c.Providers))
	for i := range c.Providers {
		if providers[i], err = secretless.LoadProvider(c.Providers[i]); err != nil {
			err = fmt.Errorf("Unable to load provider '%s' : %s", c.Providers[i].Name, err.Error())
			return
		}
	}

	subcommand.Providers = providers

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

// TODO: I am not sure what this is for
func returnStatusOfError(err error) (int, error) {
	if eerr, ok := err.(*exec.ExitError); ok {
		if ws, ok := eerr.Sys().(syscall.WaitStatus); ok {
			if ws.Exited() {
				return ws.ExitStatus(), nil
			}
		}
	}
	return 0, err
}
