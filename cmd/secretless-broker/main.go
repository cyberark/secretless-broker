package main

import (
	"flag"
	"log"

	"github.com/cyberark/secretless-broker/internal/pkg/plugin"
	yaml "gopkg.in/yaml.v1"
)

func main() {
	log.Println("Secretless starting up...")

	pluginDir := flag.String("p", "/usr/local/lib/secretless", "Directory containing Secretless plugins")
	configFile := flag.String("f", "secretless.yml", "Location of the configuration file.")
	fsWatchSwitch := flag.Bool("watch", false, "Enable automatic reloads when configuration file changes.")
	debugSwitch := flag.Bool("debug", false, "Enable debug logging.")
	flag.Parse()

	configuration := plugin.GetManager().LoadConfigurationFile(*configFile)

	if *fsWatchSwitch {
		log.Printf("Watching for changes: %s", *configFile)
		plugin.GetManager().RegisterConfigFileWatcher(*configFile)
	}

	if *debugSwitch {
		configStr, _ := yaml.Marshal(configuration)
		log.Printf("Loaded configuration: %s", configStr)
		for _, handler := range configuration.Handlers {
			handler.Debug = true
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

	plugin.GetManager().RegisterSignalHandlers(*configFile)

	plugin.GetManager().Run(configuration)
}
