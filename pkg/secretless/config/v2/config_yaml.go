package v2

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation"
	"gopkg.in/yaml.v2"
)

type configYAML struct {
	Services map[string]*serviceYAML `yaml:"services" json:"services"`
}

// Validate verifies the completeness and correctness of the configYAML.
func (c configYAML) Validate() error {
	// Validate ServiceYAML map
	errors := validation.Errors{}
	for serviceName, serviceYAML := range c.Services {
		if err := serviceYAML.Validate(); err != nil {
			errors[serviceName] = err
		}
	}
	if err := errors.Filter(); err != nil {
		return err
	}

	// Validate configYAML
	return validation.ValidateStruct(&c,
		validation.Field(&c.Services, validation.Required),
	)
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

	if err := cfgYAML.Validate(); err != nil {
		return nil, err
	}

	return cfgYAML, nil
}
