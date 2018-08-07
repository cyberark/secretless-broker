package main

import (
	"flag"
	"log"

	"github.com/conjurinc/secretless-broker/internal/pkg/plugin"
	"github.com/conjurinc/secretless-broker/pkg/secretless/config"
	yaml "gopkg.in/yaml.v1"
)

func main() {
	log.Println("Secretless starting up...")

	pluginDir := flag.String("p", "/usr/local/lib/secretless", "Directory containing Secretless plugins")
	configFile := flag.String("f", "secretless.yml", "Location of the configuration file.")
	debugSwitch := flag.Bool("debug", false, "Enable debug logging.")
	flag.Parse()

	var err error
	var configuration config.Config
	if configuration, err = config.LoadFromFile(*configFile); err != nil {
		log.Fatal(err)
	}

	if *debugSwitch {
		configStr, _ := yaml.Marshal(configuration)
		log.Printf("Loaded configuration : %s", configStr)
		for _, handler := range configuration.Handlers {
			handler.Debug = true
		}
	}

	log.Println("Loading internal plugins...")
	err = plugin.GetManager().LoadInternalPlugins(configuration)
	if err != nil {
		log.Println(err)
	}

	log.Println("Loading external library plugins...")
	err = plugin.GetManager().LoadLibraryPlugins(*pluginDir, configuration)
	if err != nil {
		log.Println(err)
	}

	plugin.GetManager().RegisterSignalHandlers()

	plugin.GetManager().Run(configuration)
}
