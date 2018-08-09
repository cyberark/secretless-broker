package example

import plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"

// Provider provides the `<ID>provider` as the value.
type Provider struct {
	Name string
}

// ProviderFactory constructs a mock Provider.
func ProviderFactory(options plugin_v1.ProviderOptions) plugin_v1.Provider {
	return &Provider{
		Name: options.Name,
	}
}

// GetName returns the name of the provider
func (provider Provider) GetName() string {
	return provider.Name
}

// GetValue returns the id + "Provider"
func (provider Provider) GetValue(id string) ([]byte, error) {
	return []byte(id + "Provider"), nil
}
