package env

import (
	"fmt"
	"os"

	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

// EnvironmentProvider provides data values from the process environment.
type EnvironmentProvider struct {
	Name string
}

// ProviderFactory constructs a EnvironmentProvider.
// No configuration or credentials are required.
func ProviderFactory(options plugin_v1.ProviderOptions) (plugin_v1.Provider, error) {
	return &EnvironmentProvider{
		Name: options.Name,
	}, nil
}

// GetName returns the name of the provider
func (p *EnvironmentProvider) GetName() string {
	return p.Name
}

// GetValues resolves multiple environment variables into values
func (p *EnvironmentProvider) GetValues(variables []config.Variable) (map[string][]byte, error) {
	result := map[string][]byte{}
	for _, variable := range variables {
		value, err := p.GetValue(variable.ID)
			if err != nil {
				return nil, err
			}
		result[variable.Name] = value
	}
	return result, nil
}

// GetValue obtains a value by ID. Any environment is a recognized ID.
func (p *EnvironmentProvider) GetValue(id string) (result []byte, err error) {
	var found bool
	envVar, found := os.LookupEnv(id)
	if found {
		result = []byte(envVar)
	} else {
		err = fmt.Errorf("%s cannot find environment variable '%s'", p.GetName(), id)
	}
	return
}
