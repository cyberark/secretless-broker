package v1

import (
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

// ProviderOptions contains the configuration for the provider instantiation
type ProviderOptions struct {
	// Name is the internal name that the provider will have. This may be different from
	// the name passed back from the provider factory.
	Name string
}

// Provider is the interface used to obtain values from a secret vault backend.
type Provider interface {
	// GetName returns the name that the Provider was instantiated with
	GetName() string

	// GetValues takes a slice of variable ids and returns their resolved values
	GetValues(variables []config.Variable) (map[string][]byte, error)

	// GetValue takes a single variable id and resolves its value
	GetValue(id string) ([]byte, error)
}
