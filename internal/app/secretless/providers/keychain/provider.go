package keychain

import (
	"fmt"
	"strings"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/app/secretless/plugin/v1"
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
