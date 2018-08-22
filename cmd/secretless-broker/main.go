package main

import (
	"flag"
	"log"
	"strings"

	yaml "gopkg.in/yaml.v1"

	"github.com/cyberark/secretless-broker/internal/pkg/plugin"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

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

	configFile := flag.String("f", "secretless.yml", "Location of the configuration file.")
	pluginDir := flag.String("p", "/usr/local/lib/secretless", "Directory containing Secretless plugins")
	configManagerSpecString := flag.String("config-mgr", "", configManagerHelp)
	fsWatchSwitch := flag.Bool("watch", false, "Enable automatic reloads when configuration file changes.")
	debugSwitch := flag.Bool("debug", false, "Enable debug logging.")
	flag.Parse()

	configuration := config.Config{}
	configManagerID, configManagerSpec := parseConfigManagerSpec(*configManagerSpecString)

	if len(*configManagerSpecString) <= 0 {
		configuration = plugin.GetManager().LoadConfigurationFile(*configFile)

		if *fsWatchSwitch {
			log.Printf("Watching for changes: %s", *configFile)
			plugin.GetManager().RegisterConfigFileWatcher(*configFile)
		}
	}

	log.Println("Loading internal plugins...")
	err := plugin.GetManager().LoadInternalPlugins(configuration)
	if err != nil {
		log.Println(err)
	}

	log.Println("Loading external library plugins...")
	err = plugin.GetManager().LoadLibraryPlugins(*pluginDir, configuration)
	if err != nil {
		log.Println(err)
	}

	if *debugSwitch {
		configStr, _ := yaml.Marshal(configuration)
		log.Printf("Loaded configuration: %s", configStr)
		for _, handler := range configuration.Handlers {
			handler.Debug = true
		}
	}

	plugin.GetManager().RegisterSignalHandlers(*configFile)
	plugin.GetManager().Run(configManagerID, configManagerSpec, configuration)
}
