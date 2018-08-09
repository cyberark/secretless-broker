package file

import (
	"io/ioutil"

	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

// Provider reads the contents of the specified file.
type Provider struct {
	Name string
}

// ProviderFactory constructs a filesystem Provider.
// No configuration or credentials are required.
func ProviderFactory(options plugin_v1.ProviderOptions) plugin_v1.Provider {
	return &Provider{
		Name: options.Name,
	}
}

// GetName returns the name of the provider
func (p Provider) GetName() string {
	return p.Name
}

// GetValue reads the contents of the identified file.
func (p Provider) GetValue(id string) ([]byte, error) {
	return ioutil.ReadFile(id)
}
