package testutils

import (
	"github.com/smartystreets/goconvey/convey"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
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
	expectedValue string) func() {
	return func() {
		values, err := provider.GetValues(id)

		convey.So(values[id].Error, convey.ShouldBeNil)
		convey.So(err, convey.ShouldBeNil)
		value := values[id]
		assertGoodProviderResponse(value, expectedValue)
	}
}

// CanProvideMultiple calls GetValues on the provider and ensures that the provider's
// responses for the each id match the expected value and there are no errors. It also
// duplicates some ids to ensure GetValues can handle multiple instances of the same id
func CanProvideMultiple(
	provider plugin_v1.Provider,
	expectedStringValueByID map[string]string,
) func() {
	return func() {
		ids := make([]string, 0, len(expectedStringValueByID)*2)
		expectedStringValueByID := map[string]string{}

		for id := range expectedStringValueByID {
			ids = append(ids, id)
			ids = append(ids, id)
		}

		responses, err := provider.GetValues(ids...)

		// Ensure no global error
		convey.So(err, convey.ShouldBeNil)
		// Ensure there many responses as there are ids
		convey.So(len(responses), convey.ShouldEqual, len(ids))
		// Ensure each id has the expected response
		for _, id := range ids {
			assertGoodProviderResponse(
				responses[id],
				expectedStringValueByID[id],
			)
		}
	}
}

// assertGoodProviderResponse asserts that a provider response has the expected string
// value and no error
func assertGoodProviderResponse(
	response plugin_v1.ProviderResponse,
	expectedValueAsStr string,
) {
	convey.So(response, convey.ShouldNotBeNil)
	convey.So(response.Error, convey.ShouldBeNil)
	convey.So(response.Value, convey.ShouldNotBeNil)
	convey.So(string(response.Value), convey.ShouldEqual, expectedValueAsStr)
}

// ReportsTestCase captures a test case where a provider is expected to return an error
type ReportsTestCase struct {
	Description       string
	ID                string
	ExpectedErrString string
}

// Reports calls GetValues on the provider and ensures that the provider response for the
// given id has the expected error and no value
func Reports(provider plugin_v1.Provider, id string, expectedErrString string) func() {
	return func() {
		values, err := provider.GetValues(id)

		convey.So(err, convey.ShouldBeNil)
		convey.So(values, convey.ShouldContainKey, id)
		convey.So(values[id].Value, convey.ShouldBeNil)
		convey.So(values[id].Error, convey.ShouldNotBeNil)
		convey.So(values[id].Error.Error(), convey.ShouldEqual, expectedErrString)
	}
}
