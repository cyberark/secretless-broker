package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValues(t *testing.T) {
	t.Run("GetValues", func(t *testing.T) {
		// ids, 4 of which are unique
		ids := []string{"foo", "err_meow", "bar", "bar", "err_meow", "err_baz"}

		t.Run("Sequentially call GetValue on unique ids", func(t *testing.T) {
			p := &MockProvider{}
			_, _ = GetValues(
				p,
				ids...,
			)

			assert.ObjectsAreEqualValues([]string{"foo", "err_meow", "bar", "err_baz"}, p.GetValueCallArgs)
		})

		t.Run("Returns good or bad responses depending on unique ids", func(t *testing.T) {
			providerResponses, err := GetValues(
				&MockProvider{},
				ids...,
			)
			assert.NoError(t, err)
			assert.Len(t, providerResponses, 4)

			ensureGoodResponse(providerResponses, "foo", t)
			ensureGoodResponse(providerResponses, "bar", t)
			ensureErrResponse(providerResponses, "err_baz", t)
			ensureErrResponse(providerResponses, "err_meow", t)
		})
	})
}

// ensureGoodResponse ensures [key] exists within a provider-responses map, and that the
// entry has no error and the value is []byte(key), in line with how
// MockProvider#GetValue works
func ensureGoodResponse(responses map[string]ProviderResponse, key string, t *testing.T) {
	assert.Contains(t, responses, key)
	assert.Nil(t, responses[key].Error)
	assert.Equal(t, key+"_value", string(responses[key].Value))
}

// ensureErrResponse ensures [key] exists within a provider-responses map, and that the
// entry has no value and the error string is equal to [key], in line with how
// MockProvider#GetValue works
func ensureErrResponse(responses map[string]ProviderResponse, key string, t *testing.T) {
	assert.Contains(t, responses, key)
	assert.Nil(t, responses[key].Value)
	assert.Equal(t, key+"_value", string(responses[key].Error.Error()))
}
