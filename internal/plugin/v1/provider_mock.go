package v1

import (
	"errors"
	"strings"
)

// MockProvider conforms to, and allows testing of, both the singleValueProvider and
// Provider interfaces
type MockProvider struct {
	GetValueCallArgs []string // keeps track of args for each call to getValue
}

// GetValue returns
// 0. If [id] has prefix 'err_', returns (nil, errors.New(id + "_value"))
// 1. Otherwise, returns ([]byte(id + "_value"), nil)
func (p *MockProvider) GetValue(id string) ([]byte, error) {
	p.GetValueCallArgs = append(p.GetValueCallArgs, id)

	if strings.HasPrefix(id, "err_") {
		return nil, errors.New(id + "_value")
	}
	return []byte(id + "_value"), nil
}

// GetValues sequentially get values for unique ids by calling GetValue
//
// If there exists any id with the prefix 'global_err_', the function will return
// (nil, errors.New(id + "_value"))
func (p *MockProvider) GetValues(ids ...string) (
	map[string]ProviderResponse,
	error,
) {
	responses := map[string]ProviderResponse{}

	for _, id := range ids {
		if _, ok := responses[id]; ok {
			continue
		}

		if strings.HasPrefix(id, "global_err_") {
			return nil, errors.New(id + "_value")
		}

		pr := ProviderResponse{}
		pr.Value, pr.Error = p.GetValue(id)

		responses[id] = pr
	}

	return responses, nil
}

// GetName simply returns "mock-provider"
func (p *MockProvider) GetName() string {
	return "mock-provider"
}
