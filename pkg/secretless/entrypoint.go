package secretless

import (
	"fmt"
	"log"
	"strings"

	"github.com/cyberark/secretless-broker/internal/pkg/plugin"
)

const programName = "secretless-broker"

const configFileManagerPluginID = "configfile"

// CLIParams objects is used to pass any CLI configuration from the initial
// entrypoint of the program to the broker's startup sequence
type CLIParams struct {
	ConfigFile          string
	ConfigManagerSpec   string
	DebugEnabled        bool
	FsWatchEnabled      bool
	PluginChecksumsFile string
	PluginDir           string
	ProfilingMode       string
	ShowVersion         bool
}

func parseConfigManagerSpec(configManagerSpecString string) (configManagerID string, configManagerSpec string) {
	if len(configManagerSpecString) == 0 {
		return "", ""
	}

	configManagerSpecItems := strings.SplitN(configManagerSpecString, "#", 2)

	if len(configManagerSpecItems) < 1 {
		log.Fatalf("ERROR: Manager config spec must be supplied in '<manager_id>[#<spec>]' form")
	}

	configManagerID = configManagerSpecItems[0]

	if len(configManagerSpecItems) > 1 {
		configManagerSpec = configManagerSpecItems[1]

	}

	return
}

// Start method is the main entry point into the broker after
// the CLI flags have been parsed
func Start(params *CLIParams, pluginManager *plugin.Manager) {
	if params.ShowVersion {
		fmt.Printf("%s v%s\n", programName, FullVersionName)
		return
	}

	log.Printf("Secretless v%s starting up...", FullVersionName)

	configManagerID, configManagerSpec := parseConfigManagerSpec(params.ConfigManagerSpec)

	// If a configuration file is specified, we don't care what config-mgr the user selected
	// as we know that we should use the configfile one.
	if len(params.ConfigFile) > 0 {
		if len(params.ConfigManagerSpec) > 0 {
			log.Printf("WARN: Config file and config manager specified" +
				" - forcing 'configfile' configuration manager!")
			configManagerID = configFileManagerPluginID
		}

		configManagerSpec = params.ConfigFile
	}

	if params.FsWatchEnabled {
		// If fsWatchSwitch is flipped but our configuration manager plugin is not 'configfile',
		// the user is requesting an impossible situation
		if configManagerID != configFileManagerPluginID {
			log.Fatalf("FATAL: Watch flag enabled on a non-filesystem based configuration" +
				"manager!")
		}

		// Attach 'watch' param as a URL query param
		configManagerSpec = configManagerSpec + fmt.Sprintf("?watch=%v", params.FsWatchEnabled)
	}

	log.Println("Loading internal plugins...")
	err := pluginManager.LoadInternalPlugins()
	if err != nil {
		log.Println(err)
	}

	log.Println("Loading external library plugins...")
	err = pluginManager.LoadLibraryPlugins(params.PluginDir, params.PluginChecksumsFile)
	if err != nil {
		log.Println(err)
	}

	// for CPU and Memory profiling
	pluginManager.SetFlags(params.ProfilingMode, params.DebugEnabled)

	pluginManager.RegisterSignalHandlers()
	pluginManager.Run(configManagerID, configManagerSpec)
}
