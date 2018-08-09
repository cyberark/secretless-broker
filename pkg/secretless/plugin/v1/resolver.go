package v1

import (
	"github.com/conjurinc/secretless-broker/pkg/secretless/config"
)

// Resolver is the interface which is used to pass a generic resolver
// down to the Listeners/Handlers.
type Resolver interface {
	// GetProvider gets back an instance of a named provider and creates it if
	// one already doesn't exist
	GetProvider(name string) (Provider, error)

	// Resolve accepts an array of variables and returns a map of resolved ones
	Resolve(variables []config.Variable) (result map[string][]byte, err error)
}
