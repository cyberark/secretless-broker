package v1

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

	// GetValue takes in an id of a variable and returns its resolved value
	GetValue(id string) ([]byte, error)
}
