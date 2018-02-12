package provider

import (
	"fmt"
	"strings"

	"github.com/conjurinc/secretless/internal/pkg/provider/keychain_provider"
)

// KeychainProvider obtains a secret from an OS keychain.
type KeychainProvider struct {
	name string
}

// NewKeychainProvider constructs a KeychainProvider.
// No configuration or credentials are required.
func NewKeychainProvider(name string) (provider Provider, err error) {
	provider = &KeychainProvider{name: name}

	return
}

// Name returns the name of the provider
func (p KeychainProvider) Name() string {
	return p.name
}

// Value reads the contents of the identified file.
func (p KeychainProvider) Value(id string) ([]byte, error) {
	tokens := strings.Split(id, "#")
	if len(tokens) != 2 {
		return nil, fmt.Errorf("Keychain secret id '%s' must be formatted as '<service>#<account>'", id)
	}

	service := tokens[0]
	account := tokens[1]

	return keychain_provider.GetGenericPassword(service, account)
}
