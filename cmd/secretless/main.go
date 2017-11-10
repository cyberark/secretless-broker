package main

import (
	"flag"
	"log"

	"github.com/kgilpin/secretless/pkg/secretless/config"
	"github.com/kgilpin/secretless/internal/app/secretless"
)

func main() {
	log.Println("Secretless starting up...")

	configFile := flag.String("config", "config.yaml", "Configuration file name")
	flag.Parse()

	configuration := config.Configure(*configFile)
	log.Printf("Loaded configuration : %v", configuration)
	p := secretless.Proxy{Config: configuration}
	p.Run()
}
