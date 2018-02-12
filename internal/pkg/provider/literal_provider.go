package provider

// LiteralProvider provides the ID as the value.
type LiteralProvider struct {
	name string
}

// NewLiteralProvider constructs a LiteralProvider.
// No configuration or credentials are required.
func NewLiteralProvider(name string) (provider Provider, err error) {
	provider = &LiteralProvider{name: name}

	return
}

// Name returns the name of the provider
func (p LiteralProvider) Name() string {
	return p.name
}

// Value returns the id.
func (p LiteralProvider) Value(id string) ([]byte, error) {
	return []byte(id), nil
}
