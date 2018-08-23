package literal

import (
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)


// Provider provides the ID as the value.
type Provider struct {
	Name string
}

// ProviderFactory constructs a literal value Provider.
// No configuration or credentials are required.
func ProviderFactory(options plugin_v1.ProviderOptions) (plugin_v1.Provider, error) {
	return &Provider{
		Name: options.Name,
	}, nil
}

// GetName returns the name of the provider
func (p *Provider) GetName() string {
	return p.Name
}

// GetValues resolves multiple literal fields into values
func (p *Provider) GetValues(variables []config.Variable) (map[string][]byte, error) {
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

// GetValue returns the id.
func (p *Provider) GetValue(id string) ([]byte, error) {
	return []byte(id), nil
}
