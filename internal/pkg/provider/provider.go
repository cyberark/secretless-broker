package provider

import (
	"fmt"
	"sync"

	"github.com/conjurinc/secretless/pkg/secretless"
)

// NewProvider creates a named provider.
type NewProvider func(name string) (secretless.Provider, error)

var providers = make(map[string]secretless.Provider)

// GetProvider finds or creates a named provider.
func GetProvider(name string) (provider secretless.Provider, err error) {
	var mutex = &sync.Mutex{}

	mutex.Lock()
	defer mutex.Unlock()

	if provider = providers[name]; provider != nil {
		return
	}

	var factory NewProvider
	if factory, err = newProvider(name); err != nil {
		return
	}

	if provider, err = factory(name); err != nil {
		return
	}

	providers[name] = provider

	return
}

func newProvider(name string) (provider NewProvider, err error) {
	switch name {
	case "env", "environment":
		provider = NewEnvironmentProvider
	case "literal":
		provider = NewLiteralProvider
	case "file":
		provider = NewFileProvider
	case "conjur":
		provider = NewConjurProvider
	case "vault":
		provider = NewVaultProvider
	case "keychain":
		provider = NewKeychainProvider
	default:
		err = fmt.Errorf("Unrecognized provider type '%s'", name)
	}
	return
}
