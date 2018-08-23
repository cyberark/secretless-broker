package file

import (
	"io/ioutil"

	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

// Provider reads the contents of the specified file.
type Provider struct {
	Name string
}

// ProviderFactory constructs a filesystem Provider.
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

// GetValues resolves multiple files into values
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

// GetValue reads the contents of the identified file.
func (p *Provider) GetValue(id string) ([]byte, error) {
	return ioutil.ReadFile(id)
}
