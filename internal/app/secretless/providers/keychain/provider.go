package keychain

import (
	"fmt"
	"strings"

	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

// Provider obtains a secret from an OS keychain.
type Provider struct {
	Name string
}

// ProviderFactory constructs a keychain Provider.
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

// GetValues resolves multiple keychain variables into values
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
	tokens := strings.Split(id, "#")
	if len(tokens) != 2 {
		return nil, fmt.Errorf("Keychain secret id '%s' must be formatted as '<service>#<account>'", id)
	}

	service := tokens[0]
	account := tokens[1]

	return GetGenericPassword(service, account)
}
