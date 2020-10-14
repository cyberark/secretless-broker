package v1

// ProviderOptions contains the configuration for the provider instantiation
type ProviderOptions struct {
	// Name is the internal name that the provider will have. This may be different from
	// the name passed back from the provider factory.
	Name string
}

// ProviderResponse is the response from the provider for a given secret request
type ProviderResponse struct {
	Value []byte
	Error error
}

// Provider is the interface used to obtain values from a secret vault backend.
type Provider interface {
	// GetName returns the name that the Provider was instantiated with
	GetName() string

	// GetValues takes in variable ids and returns their resolved values
	GetValues(ids ...string) (map[string]ProviderResponse, error)
}

type singleValueProvider interface {
	// GetValue takes in an id of a variable and returns its resolved value
	GetValue(id string) ([]byte, error)
}

// GetValues takes in variable ids and returns their resolved values by making sequential
// getValueCallArgs to a singleValueProvider.
// This is a convenience function since most providers with batch retrieval capabilities
// will have need the exact same code. Note: most internal providers simply use this
// function in their implementation of the Provider interface's GetValues method.
func GetValues(
	p singleValueProvider,
	ids ...string,
) (map[string]ProviderResponse, error) {
	responses := map[string]ProviderResponse{}

	for _, id := range ids {
		if _, ok := responses[id]; ok {
			continue
		}

		pr := ProviderResponse{}
		pr.Value, pr.Error = p.GetValue(id)

		responses[id] = pr
	}

	return responses, nil
}
