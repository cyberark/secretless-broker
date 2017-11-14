package config

import (
  "io/ioutil"
  "log"
  "regexp"

  "gopkg.in/yaml.v2"
)

type Keychain struct {
  Service  string
  Username string
}

type ValueFrom struct {
  Conjur      string
  Environment string
  File        string
  Keychain    Keychain
}

type Variable struct {
  Name      string
  Value     string
  ValueFrom ValueFrom `yaml:"value_from"`
}

type Authorization struct {
  None      bool
  Conjur    string
  Passwords map[string]string
}

type Listener struct {
  Name     string
  Protocol string
  Address  string
  Socket   string
}

type Handler struct {
  Name          string
  Type          string
  Listener      string
  Authorization Authorization
  Debug         bool
  Match         []string
  Patterns      []*regexp.Regexp
  Backend       []Variable
}

type Config struct {
  Listeners []Listener
  Handlers  []Handler
}

func Configure(fileName string) Config {
	var err error

  config := Config{}

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

  return config
}
