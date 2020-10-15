package v1

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetValues(t *testing.T) {
	Convey("GetValues", t, func() {
		// ids, 4 of which are unique
		ids := []string{"foo", "err_meow", "bar", "bar", "err_meow", "err_baz"}

		Convey("Sequentially call GetValue on unique ids", func() {
			p := &MockProvider{}
			_, _ = GetValues(
				p,
				ids...,
			)

			So(p.GetValueCallArgs, ShouldResemble, []string{"foo", "err_meow", "bar", "err_baz"})
		})

		Convey("Returns good or bad responses depending on unique ids", func() {
			providerResponses, err := GetValues(
				&MockProvider{},
				ids...,
			)
			So(err, ShouldBeNil)
			So(len(providerResponses), ShouldEqual, 4)

			ensureGoodResponse(providerResponses, "foo")
			ensureGoodResponse(providerResponses, "bar")
			ensureErrResponse(providerResponses, "err_baz")
			ensureErrResponse(providerResponses, "err_meow")
		})
	})
}

// ensureGoodResponse ensures [key] exists within a provider-responses map, and that the
// entry has no error and the value is []byte(key), in line with how
// MockProvider#GetValue works
func ensureGoodResponse(responses map[string]ProviderResponse, key string) {
	So(responses, ShouldContainKey, key)
	So(responses[key].Error, ShouldBeNil)
	So(responses[key].Value, ShouldResemble, []byte(key+"_value"))
}

// ensureErrResponse ensures [key] exists within a provider-responses map, and that the
// entry has no value and the error string is equal to [key], in line with how
// MockProvider#GetValue works
func ensureErrResponse(responses map[string]ProviderResponse, key string) {
	So(responses, ShouldContainKey, key)
	So(responses[key].Value, ShouldBeNil)
	So(responses[key].Error.Error(), ShouldEqual, key+"_value")
}
