package testutils

import (
	"testing"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/stretchr/testify/assert"
)

// CanProvideTestCase captures a test case where a provider is expected to return a value
// and no error
type CanProvideTestCase struct {
	Description   string
	ID            string
	ExpectedValue string
}

// CanProvide calls GetValues on the provider and ensures that the provider response for
// the given id has the expected value and no error
func CanProvide(provider plugin_v1.Provider,
	id string,
	expectedValue string) func(t *testing.T) {
	return func(t *testing.T) {
		values, err := provider.GetValues(id)

		assert.NoError(t, values[id].Error)
		assert.NoError(t, err)
		value := values[id]
		assertGoodProviderResponse(value, expectedValue, t)
	}
}

// CanProvideMultiple calls GetValues on the provider and ensures that the provider's
// responses for the each id match the expected value and there are no errors. It also
// duplicates some ids to ensure GetValues can handle multiple instances of the same id
func CanProvideMultiple(
	provider plugin_v1.Provider,
	expectedStringValueByID map[string]string,
) func(t *testing.T) {
	return func(t *testing.T) {
		ids := make([]string, 0, len(expectedStringValueByID)*2)
		expectedStringValueByID := map[string]string{}

		for id := range expectedStringValueByID {
			ids = append(ids, id)
			ids = append(ids, id)
		}

		responses, err := provider.GetValues(ids...)

		// Ensure no global error
		assert.NoError(t, err)
		// Ensure there many responses as there are ids
		assert.Len(t, responses, len(ids))
		// Ensure each id has the expected response
		for _, id := range ids {
			assertGoodProviderResponse(
				responses[id],
				expectedStringValueByID[id],
				t,
			)
		}
	}
}

// assertGoodProviderResponse asserts that a provider response has the expected string
// value and no error
func assertGoodProviderResponse(
	response plugin_v1.ProviderResponse,
	expectedValueAsStr string,
	t *testing.T,
) {
	assert.NotNil(t, response)
	assert.NoError(t, response.Error)
	assert.NotNil(t, response.Value)
	assert.Equal(t, expectedValueAsStr, string(response.Value))
}

// ReportsTestCase captures a test case where a provider is expected to return an error
type ReportsTestCase struct {
	Description       string
	ID                string
	ExpectedErrString string
}

// Reports calls GetValues on the provider and ensures that the provider response for the
// given id has the expected error and no value
func Reports(provider plugin_v1.Provider, id string, expectedErrString string) func(t *testing.T) {
	return func(t *testing.T) {
		values, err := provider.GetValues(id)

		assert.NoError(t, err)
		assert.Contains(t, values, id)
		assert.Nil(t, values[id].Value)
		assert.Error(t, values[id].Error)
		assert.EqualError(t, values[id].Error, expectedErrString)
	}
}
