package main

import (
	"flag"
	"log"

	"github.com/kgilpin/secretless-pg/config"
	"github.com/kgilpin/secretless-pg/proxy"
)

func main() {
	log.Println("Secretless starting up...")

	configFile := flag.String("config", "config.yaml", "Configuration file name")
	flag.Parse()

	configuration := config.Configure(*configFile)
	log.Printf("Loaded configuration : %v", configuration)
	p := proxy.Proxy{Config: configuration}
	p.Run()
}
