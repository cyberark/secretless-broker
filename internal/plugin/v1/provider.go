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

	// GetValues takes in variable ids and returns their resolved values
	GetValues(ids ...string) ([][]byte, error)
}

type singleValueProvider interface {
	// GetValue takes in an id of a variable and returns its resolved value
	GetValue(id string) ([]byte, error)
}

// GetValues takes in variable ids and returns their resolved values by making sequential
// calls to a singleValueProvider.
// This is a convenience function since most providers with batch retrieval capabilities
// will have need the exact same code. Note: most internal providers simply use this
// function in their implementation of the Provider interface's GetValues method.
func GetValues(
	p singleValueProvider,
	ids ...string,
) ([][]byte, error) {
	var err error
	var res = make([][]byte, len(ids))

	for idx, id := range ids {
		res[idx], err = p.GetValue(id)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
