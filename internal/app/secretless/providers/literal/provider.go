package literal

import plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"

// Provider provides the ID as the value.
type Provider struct {
	Name string
}

// ProviderFactory constructs a literal value Provider.
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

// GetValue returns the id.
func (p Provider) GetValue(id string) ([]byte, error) {
	return []byte(id), nil
}
