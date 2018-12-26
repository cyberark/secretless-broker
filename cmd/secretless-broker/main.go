package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/cyberark/secretless-broker/internal/pkg/plugin"
)

const configFileManagerPluginID = "configfile"

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

func main() {
	log.Println("Secretless starting up...")

	configManagerHelp := "(Optional) Specify a config manager ID and an optional manager-specific spec string "
	configManagerHelp += "(eg '<name>[#<filterSpec>]'). "
	configManagerHelp += "Default will try to use 'secretless.yml' configuration."

	configFile := flag.String("f", "", "Location of the configuration file.")

	// for CPU and Memory profiling
	// Acceptable values to input: cpu or memory
	profileSwitch := flag.String("profile", "", "Enable and set the profiling mode to the value provided. Acceptable values are 'cpu' or 'memory'.")

	configManagerSpecString := flag.String("config-mgr", "configfile", configManagerHelp)
	debugSwitch := flag.Bool("debug", false, "Enable debug logging.")
	fsWatchSwitch := flag.Bool("watch", false, "Enable automatic reloads when configuration file changes.")
	pluginDir := flag.String("p", "/usr/local/lib/secretless", "Directory containing Secretless plugins")
	flag.Parse()

	configManagerID, configManagerSpec := parseConfigManagerSpec(*configManagerSpecString)

	// If a configuration file is specified, we don't care what config-mgr the user selected
	// as we know that we should use the configfile one.
	if len(*configFile) > 0 {
		if len(*configManagerSpecString) > 0 {
			log.Printf("WARN: Config file and config manager specified" +
				" - forcing 'configfile' configuration manager!")
			configManagerID = configFileManagerPluginID
		}

		configManagerSpec = *configFile
	}

	if *fsWatchSwitch {
		// If fsWatchSwitch is flipped but our configuration manager plugin is not 'configfile',
		// the user is requesting an impossible situation
		if configManagerID != configFileManagerPluginID {
			log.Fatalf("FATAL: Watch flag enabled on a non-filesystem based configuration" +
				"manager!")
		}

		// Attach 'watch' param as a URL query param
		configManagerSpec = configManagerSpec + fmt.Sprintf("?watch=%v", *fsWatchSwitch)
	}

	log.Println("Loading internal plugins...")
	err := plugin.GetManager().LoadInternalPlugins()
	if err != nil {
		log.Println(err)
	}

	log.Println("Loading external library plugins...")
	err = plugin.GetManager().LoadLibraryPlugins(*pluginDir)
	if err != nil {
		log.Println(err)
	}

	// for CPU and Memory profiling
	plugin.GetManager().SetFlags(*profileSwitch, *debugSwitch)

	plugin.GetManager().RegisterSignalHandlers()
	plugin.GetManager().Run(configManagerID, configManagerSpec)
}
