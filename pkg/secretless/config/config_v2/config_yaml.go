package config_v2

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

type configYAML struct {
	Services map[string]*serviceYAML
}

type serviceYAML struct {
	Protocol    string          `yaml:"protocol" json:"protocol"`
	ListenOn    string          `yaml:"listenOn" json:"listenOn"`
	Credentials credentialsYAML `yaml:"credentials" json:"credentials"`
	Config      interface{}     `yaml:"config" json:"config"`
}

// CredentialYAML needs to be an interface because it contains arbitrary YAML
// dictionaries, since end user credential names can be anything.

type credentialsYAML map[string]interface{}

func newConfigYAML(rawYAML []byte) (*configYAML, error) {
	if len(rawYAML) == 0 {
		return nil, fmt.Errorf("empty file contents given to NewConfig")
	}
	cfgYAML := &configYAML{}
	err := yaml.Unmarshal(rawYAML, cfgYAML)
	if err != nil {
		return nil, err
	}

	return cfgYAML, nil
}

