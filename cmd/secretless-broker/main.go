package main

import (
	"flag"

	"github.com/cyberark/secretless-broker/pkg/secretless/entrypoint"
)

func main() {
	params := CmdLineParams()
	entrypoint.StartSecretless(params)
}

// CmdLineParams parses all cmd line options and returns the resulting SecretlessOptions.
func CmdLineParams() *entrypoint.SecretlessOptions {
	configManagerHelp := "(Optional) Specify a config manager ID and an optional manager-specific spec string "
	configManagerHelp += "(eg '<name>[#<filterSpec>]'). "
	configManagerHelp += "Default will try to use 'secretless.yml' configuration."

	params := entrypoint.SecretlessOptions{}

	flag.StringVar(&params.ConfigFile, "f", "", "Location of the configuration file.")

	// for CPU and Memory profiling
	// Acceptable values to input: cpu or memory
	flag.StringVar(&params.ProfilingMode, "profile", "",
		"Enable and set the profiling mode to the value provided. Acceptable values are 'cpu' or 'memory'.")

	// For development use only; enable more verbose debug logging
	flag.BoolVar(&params.DebugEnabled, "debug", false, "Enable debug logging.")

	flag.StringVar(&params.ConfigManagerSpec, "config-mgr", "configfile", configManagerHelp)
	flag.BoolVar(&params.FsWatchEnabled, "watch", false,
		"Enable automatic reloads when configuration file changes.")
	flag.StringVar(&params.PluginDir, "p", "/usr/local/lib/secretless",
		"Directory containing Secretless plugins")
	flag.StringVar(&params.PluginChecksumsFile, "s", "",
		"Path to a file of sha256sum plugin checksums")

	// Flag.parse only covers `-version` flag but for `version`, we need to explicitly
	// check the args
	showVersion := flag.Bool("version", false, "Show current version")

	flag.Parse()

	// Either the flag or the arg should be enough to show the version
	params.ShowVersion = *showVersion || flag.Arg(0) == "version"

	return &params
}
