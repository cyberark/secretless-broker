package v1

import (
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

// Resolver is the interface which is used to pass a generic resolver
// down to the Listeners/Handlers.
type Resolver interface {
	// Provider gets back an instance of a named provider and creates it if
	// one already doesn't exist
	Provider(name string) (Provider, error)

	// Resolve accepts an array of credentials and returns a map of resolved ones
	Resolve(credentials []*config_v2.Credential) (result map[string][]byte, err error)
}
