package config

import (
	"io/ioutil"
	"log"
	"regexp"

	"gopkg.in/yaml.v2"
)

type KeychainValue struct {
	Service  string
	Username string
}

type ValueFrom struct {
	Literal     string
	Environment string
	File        string
	Keychain    KeychainValue `yaml:"keychain"`
	Provider    string
	Id          string
}

type Variable struct {
	Name  string
	Value ValueFrom `yaml:"value"`
}

type Provider struct {
	Name          string
	Type          string
	Configuration []Variable
	Credentials   []Variable
}

type Listener struct {
	Name        string
	Protocol    string
	Address     string
	Socket      string
	CACertFiles []string `yaml:"caCertFiles"`
}

type Handler struct {
	Name        string
	Type        string
	Listener    string
	Debug       bool
	Match       []string
	Patterns    []*regexp.Regexp
	Credentials []Variable
}

type Config struct {
	Providers []Provider
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
