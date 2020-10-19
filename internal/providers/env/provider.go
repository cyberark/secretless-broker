package env

import (
	"fmt"
	"os"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
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

// GetValues takes in variable ids and returns their resolved values. This method is
// needed to the Provider interface
func (p *EnvironmentProvider) GetValues(ids ...string) (map[string]plugin_v1.ProviderResponse, error) {
	return plugin_v1.GetValues(p, ids...)
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
