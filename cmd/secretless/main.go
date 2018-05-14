package main

import (
	"flag"
	"log"

	"github.com/conjurinc/secretless/internal/app/secretless"
	"github.com/conjurinc/secretless/internal/pkg/plugin"
	"github.com/conjurinc/secretless/pkg/secretless/config"
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

	err = plugin.GetManager().LoadPlugins(*pluginDir, configuration)
	if err != nil {
		log.Println(err)
	}

	p := secretless.Proxy{Config: configuration}
	p.Run()
}
