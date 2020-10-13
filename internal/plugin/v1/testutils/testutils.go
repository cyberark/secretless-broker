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
func CanProvide(provider plugin_v1.Provider, id string, expectedValue string) func() {
	return func() {
		values, err := provider.GetValues(id)

		convey.So(err, convey.ShouldBeNil)
		convey.So(values[id], convey.ShouldNotBeNil)
		convey.So(values[id].Error, convey.ShouldBeNil)
		convey.So(values[id].Value, convey.ShouldNotBeNil)
		convey.So(string(values[id].Value), convey.ShouldEqual, expectedValue)
	}
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
