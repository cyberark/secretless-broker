package provider

// Provider is the interface used to obtain values from a secret vault backend.
type Provider interface {
	Name() string
	Value(id string) ([]byte, error)
}
