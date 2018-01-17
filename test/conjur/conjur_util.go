package main

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v1"
)

// ConjurConfig holds the configuration that is created by start.sh
type ConjurConfig struct {
	URL     string
	Account string
	APIKey  string `yaml:"api_key"`
}

// LoadTestConjurConfig provides a means for running a native Go environment with
// Conjur running in a container.
func LoadTestConjurConfig() ConjurConfig {
	var err error

	conjurrcFile := "./tmp/.conjurrc"

	_, err = os.Stat(conjurrcFile)
	if os.IsNotExist(err) {
		panic(fmt.Sprintf("conjurrc file %s does not exist; run ./start.sh to create it", conjurrcFile))
	}

	conjurConfig := ConjurConfig{}
	buf, err := ioutil.ReadFile(conjurrcFile)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(buf, &conjurConfig)
	if err != nil {
		panic(err)
	}

	url := os.Getenv("CONJUR_APPLIANCE_URL")
	if url != "" {
		conjurConfig.URL = url
	}

	return conjurConfig
}
