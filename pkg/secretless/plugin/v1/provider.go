package v1

// ProviderOptions contains the configuration for the provider instantiation
type ProviderOptions struct {
	Name string
}

// Provider is the interface used to obtain values from a secret vault backend.
type Provider interface {
	GetName() string
	GetValue(id string) ([]byte, error)
}
