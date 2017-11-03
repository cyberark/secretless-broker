package config

import (
  "io/ioutil"
  "log"

  "gopkg.in/yaml.v2"
)

type ValueFrom struct {
  Conjur      string
  Environment string
  File        string
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

type ListenerConfig struct {
  Address       string
  Socket        string
  Authorization Authorization
  Debug         bool
  Backend       []Variable
}

type Listener struct {
  Name   string
  Type   string
  Config ListenerConfig `yaml:"configuration"`
}

type Config struct {
  Listeners []Listener
}

func Configure(fileName string) Config {
  config := Config{}

  buffer, err := ioutil.ReadFile(fileName)
  if err != nil {
    log.Fatalf("Unable to read config file %s: %s", fileName, err)
  }
  err = yaml.Unmarshal(buffer, &config)
  if err != nil {
    log.Fatalf("Unable to load config file %s : %s", fileName, err)
  }

  return config
}
