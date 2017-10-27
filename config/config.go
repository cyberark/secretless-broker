package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type BackendConfig struct {
	Address  string
	Username string
	Password string
	Database string
	Options  map[string]string
}

type Authorization struct {
	None     bool
	Resource string
	Users    map[string]string `yaml:"authorized_users"`
}

type Config struct {
	Address         string
	Socket          string
	Authorization   Authorization
	Backend         BackendConfig     `yaml:"backend"`
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
