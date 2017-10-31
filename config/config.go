package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type ValueFrom struct {
	Conjur string
	File   string
}

type Variable struct {
	Name  string
	Value string
	ValueFrom ValueFrom `yaml:"value_from"`
}

type Authorization struct {
	None      bool
	Conjur    string
	Passwords map[string]string
}

type Config struct {
	Address         string
	Socket          string
	Authorization   Authorization
	Backend         []Variable
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
