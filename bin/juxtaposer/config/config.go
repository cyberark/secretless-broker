package config

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v3"

	formatter_api "github.com/cyberark/secretless-broker/bin/juxtaposer/formatter/api"
)

// Config is the main structure used to define the perfagent parameters
type Config struct {
	Backends   map[string]Backend                        `yaml:"backends"`
	Comparison Comparison                                `yaml:"comparison"`
	Driver     string                                    `yaml:"driver"`
	Formatters map[string]formatter_api.FormatterOptions `yaml:"formatters"`
}

type Backend struct {
	Database    string `yaml:"database"`
	Debug       bool   `yaml:"debug"`
	Description string `yaml:"description"`
	Host        string `yaml:"host"`
	Ignore      bool   `yaml:"ignore"`
	Password    string `yaml:"password"`
	Port        string `yaml:"port"`
	SslMode     string `yaml:"sslmode"`
	Socket      string `yaml:"socket"`
	Username    string `yaml:"username"`
}

type Comparison struct {
	BaselineBackend             string `yaml:"baselineBackend"`
	BaselineMaxThresholdPercent int    `yaml:"baselineMaxThresholdPercent"`
	Rounds                      string `yaml:"rounds"`
	Silent                      bool   `yaml:"silent"`
	Style                       string `yaml:"style"`
	Type                        string `yaml:"type"`
}

func (configuration *Config) verify() error {
	if !strings.HasPrefix(configuration.Comparison.Type, "sql/") {
		return fmt.Errorf("comparison type not supported: %s", configuration.Comparison.Type)
	}

	connectionType := configuration.Comparison.Type[4:]
	if connectionType != "persistent" &&
		connectionType != "recreate" {
		return fmt.Errorf("comparison connection type not supported: %s",
			connectionType)
	}

	if configuration.Comparison.Style != "select" {
		return fmt.Errorf("comparison style supported: %s", configuration.Comparison.Style)
	}

	if len(configuration.Formatters) == 0 {
		return fmt.Errorf("no formatters defined")
	}

	baselineBackend := configuration.Comparison.BaselineBackend
	if baselineBackend == "" {
		return fmt.Errorf("comparison baselineBackend must be specified")
	}

	if _, ok := configuration.Backends[baselineBackend]; !ok {
		return fmt.Errorf("comparison baseline backend '%s' not found",
			baselineBackend)
	}

	return nil
}

func NewConfiguration(configFile string) (*Config, error) {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// Default options
	configuration := Config{
		Comparison: Comparison{
			BaselineMaxThresholdPercent: 120,
			Rounds:                      "1000",
			Style:                       "select",
			Type:                        "sql/persistent",
		},
		Formatters: map[string]formatter_api.FormatterOptions{
			"stdout": formatter_api.FormatterOptions{},
		},
	}
	err = yaml.Unmarshal(yamlFile, &configuration)
	if err != nil {
		return nil, err
	}

	// Slice out any backends which are ignored
	filteredBackends := map[string]Backend{}
	for backendName, backendConfig := range configuration.Backends {
		if backendConfig.Ignore == false {
			filteredBackends[backendName] = backendConfig
		}
	}

	configuration.Backends = filteredBackends

	err = configuration.verify()
	if err != nil {
		return nil, err
	}

	return &configuration, nil
}
