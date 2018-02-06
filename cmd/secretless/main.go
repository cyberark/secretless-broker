package main

import (
	"flag"
	"log"

	"github.com/conjurinc/secretless/internal/app/secretless"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	yaml "gopkg.in/yaml.v1"
)

func main() {
	log.Println("Secretless starting up...")

	configFile := flag.String("config", "config.yaml", "Configuration file name")
	debugSwitch := flag.Bool("debug", false, "Print debug information")
	flag.Parse()

	var configuration config.Config
	if *configFile != "" {
		configuration = config.Configure(*configFile)
	}

	if *debugSwitch {
		configStr, _ := yaml.Marshal(configuration)
		log.Printf("Loaded configuration : %s", configStr)
		for _, handler := range configuration.Handlers {
			handler.Debug = true
		}
	}

	p := secretless.Proxy{Config: configuration}
	p.Run()
}
