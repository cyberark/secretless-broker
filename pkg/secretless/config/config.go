package config

import (
	"io/ioutil"
	"log"
	"regexp"

	"gopkg.in/yaml.v2"
)

// Variable is a named secret.
type Variable struct {
	// Name is the name by which the variable will be used by the client.
	Name string
	// Provider is the provider name.
	Provider string
	// Value is the identifier of the secret that the Provider will load.
	ID string
}

// Listener listens on a port on socket for inbound connections, which are
// handed off to Handlers.
type Listener struct {
	Name        string
	Protocol    string
	Address     string
	Socket      string
	CACertFiles []string `yaml:"caCertFiles"`
}

// Handler processes an inbound message and connects to a specified backend
// using Credentials which it fetches from a provider.
type Handler struct {
	Name        string
	Type        string
	Listener    string
	Debug       bool
	Match       []string
	Patterns    []*regexp.Regexp
	Credentials []Variable
}

// Config is the main configuration structure for Secretless.
// It lists and configures the protocol listeners and handlers.
type Config struct {
	Listeners []Listener
	Handlers  []Handler
}

// Configure loads Config data from the specified filename. The file must
// exist, or the program with panic.
func Configure(fileName string) (config Config) {
	var err error

	buffer, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Unable to read config file %s: %s", fileName, err)
	}
	err = yaml.Unmarshal(buffer, &config)
	if err != nil {
		log.Fatalf("Unable to load config file %s : %s", fileName, err)
	}

	for i := range config.Handlers {
		handler := &config.Handlers[i]
		handler.Patterns = make([]*regexp.Regexp, len(handler.Match))
		for i, pattern := range handler.Match {
			pattern, err := regexp.Compile(pattern)
			if err != nil {
				panic(err.Error())
			} else {
				handler.Patterns[i] = pattern
			}
		}
	}

	return
}
