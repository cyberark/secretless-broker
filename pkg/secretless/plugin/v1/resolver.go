package v1

import "github.com/conjurinc/secretless/pkg/secretless/config"

// Resolver is the interface which is used to pass a generic resolver
// down to the Listeners/Handlers.
type Resolver interface {
	Resolve(variables []config.Variable) (result map[string][]byte, err error)
}
