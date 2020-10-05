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
		id := "accessToken"
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldBeNil)
		So(values[id].Value, ShouldNotBeNil)

		token := make(map[string]string)
		err = json.Unmarshal(values[id].Value, &token)
		So(err, ShouldBeNil)
		So(token["protected"], ShouldNotBeNil)
		So(token["payload"], ShouldNotBeNil)
	})

	Convey("Can provide a secret to a fully qualified variable", t, func() {
		id := "dev:variable:db/password"
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldBeNil)
		So(values[id].Value, ShouldNotBeNil)
		So(string(values[id].Value), ShouldEqual, "secret")
	})

	Convey("Can retrieve a secret value with spaces", t, func() {
		id := "my var"
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldBeNil)
		So(values[id].Value, ShouldNotBeNil)
		So(string(values[id].Value), ShouldEqual, "othersecret")
	})

	Convey("Can provide the default Conjur account name", t, func() {
		id := "variable:db/password"
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldBeNil)
		So(values[id].Value, ShouldNotBeNil)
		So(string(values[id].Value), ShouldEqual, "secret")
	})

	Convey("Can provide the default Conjur account name and resource type", t, func() {
		id := "db/password"
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldBeNil)
		So(values[id].Value, ShouldNotBeNil)
		So(string(values[id].Value), ShouldEqual, "secret")
	})

	Convey("Cannot provide an unknown value", t, func() {
		id := "foobar"

		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldNotBeNil)
		So(values[id].Error.Error(), ShouldEqual, "404 Not Found. Variable 'foobar' not found in account 'dev'.")
		So(values[id].Value, ShouldBeNil)
	})
}
