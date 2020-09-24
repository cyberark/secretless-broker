package main

import (
	"encoding/json"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	. "github.com/smartystreets/goconvey/convey"

	pluginV1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/providers"
)

// TestConjur_Provider tests the ability of the ConjurProvider to provide a Conjur accessToken
// as well as secret values.
func TestConjur_Provider(t *testing.T) {
	var err error
	var provider pluginV1.Provider
	name := "conjur"

	options := pluginV1.ProviderOptions{
		Name: name,
	}

	Convey("Can create the Conjur provider", t, func() {
		provider, err = providers.ProviderFactories[name](options)
		So(err, ShouldBeNil)
	})

	Convey("Has the expected provider name", t, func() {
		So(provider.GetName(), ShouldEqual, "conjur")
	})

	Convey("Can provide an access token", t, func() {
		values, err := provider.GetValues("accessToken")
		So(err, ShouldBeNil)

		token := make(map[string]string)
		err = json.Unmarshal(values[0], &token)
		So(err, ShouldBeNil)
		So(token["protected"], ShouldNotBeNil)
		So(token["payload"], ShouldNotBeNil)
	})

	Convey("Can provide a secret to a fully qualified variable", t, func() {
		values, err := provider.GetValues("dev:variable:db/password")
		So(err, ShouldBeNil)

		So(string(values[0]), ShouldEqual, "secret")
	})

	Convey("Can retrieve a secret value with spaces", t, func() {
		values, err := provider.GetValues("my var")
		So(err, ShouldBeNil)

		So(string(values[0]), ShouldEqual, "othersecret")
	})

	Convey("Can provide the default Conjur account name", t, func() {
		values, err := provider.GetValues("variable:db/password")
		So(err, ShouldBeNil)

		So(string(values[0]), ShouldEqual, "secret")
	})

	Convey("Can provide the default Conjur account name and resource type", t, func() {
		values, err := provider.GetValues("db/password")
		So(err, ShouldBeNil)

		So(string(values[0]), ShouldEqual, "secret")
	})

	Convey("Cannot provide an unknown value", t, func() {
		_, err = provider.GetValues("foobar")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "404 Not Found. Variable 'foobar' not found in account 'dev'.")
	})
}
