package example

import (
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

// Provider provides the `<ID>provider` as the value.
type Provider struct {
	Name string
}

// ProviderFactory constructs a mock Provider.
func ProviderFactory(options plugin_v1.ProviderOptions) (plugin_v1.Provider, error) {
	return &Provider{
		Name: options.Name,
	}, nil
}

// GetName returns the name of the provider
func (provider *Provider) GetName() string {
	return provider.Name
}

// GetValues returns multiple values
func (provider *Provider) GetValues(variables []config.Variable) (map[string][]byte, error) {
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

// GetValue returns the id + "Provider"
func (provider *Provider) GetValue(id string) ([]byte, error) {
	return []byte(id + "Provider"), nil
}
