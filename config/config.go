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
	Resource string
}

type Config struct {
	Address         string
	Authorization   Authorization
	AuthorizedUsers map[string]string `yaml:"authorized_users"`
	Backend         BackendConfig     `yaml:"backend"`
}

func Configure(fileName string) Config {
	config := Config{Address: "localhost:5432"}
	config.Backend = BackendConfig{}

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
