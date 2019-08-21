package v2

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type configYAML struct {
	Services map[string]*serviceYAML
}

type serviceYAML struct {
	// Protocol specifies the service connector by protocol.
	// It is an internal detail.
	//
	// Deprecated: Protocol exists for historical compatibility
	// and should not be used. To specify the service connector,
	// use the Connector field.
	Protocol    string          `yaml:"protocol" json:"protocol"`
	Connector   string          `yaml:"connector" json:"connector"`
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
