package main

import (
	"testing"

	_ "github.com/joho/godotenv/autoload"
	. "github.com/smartystreets/goconvey/convey"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/providers"
)

func TestVault_Provider(t *testing.T) {
	var err error
	var provider plugin_v1.Provider
	name := "vault"

	options := plugin_v1.ProviderOptions{
		Name: name,
	}

	Convey("Can create the Vault provider", t, func() {
		provider, err = providers.ProviderFactories[name](options)
		So(err, ShouldBeNil)
	})

	Convey("Has the expected provider name", t, func() {
		So(provider.GetName(), ShouldEqual, "vault")
	})

	Convey("Reports when the secret is not found", t, func() {
		id := "foobar"
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldNotBeNil)
		So(values[id].Error, ShouldEqual, "HashiCorp Vault provider could not find secret 'foobar'")
		So(values[id].Value, ShouldBeNil)
	})

	Convey("Reports when a field in the secret is not found", t, func() {
		id := "cubbyhole/first-secret#foo.bar"
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldNotBeNil)
		So(values[id].Error, ShouldEqual, "HashiCorp Vault provider expects secret in 'foo.bar' at 'cubbyhole/first-secret'")
		So(values[id].Value, ShouldBeNil)
	})

	Convey("Can provide a cubbyhole secret", t, func() {
		id := "cubbyhole/first-secret#some-key"
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldBeNil)
		So(values[id].Value, ShouldNotBeNil)
		So(string(values[id].Value), ShouldEqual, "one")
	})

	Convey("Can provide a cubbyhole secret with default field name", t, func() {
		id :="cubbyhole/second-secret"
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldBeNil)
		So(values[id].Value, ShouldNotBeNil)
		So(string(values[id].Value), ShouldEqual, "two")
	})

	Convey("Can provide a KV v1 secret", t, func() {
		id := "kv/db/password#password"
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldBeNil)
		So(values[id].Value, ShouldNotBeNil)
		So(string(values[id].Value), ShouldEqual, "db-secret")
	})

	Convey("Can provide a KV v1 secret with default field name", t, func() {
		id := "kv/web/password"
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldBeNil)
		So(values[id].Value, ShouldNotBeNil)
		So(string(values[id].Value), ShouldEqual, "web-secret")
	})

	// note the "data" in path and in the fields to navigate, which is required in KV v2
	Convey("Can provide latest KV v2 secret", t, func() {
		id := "secret/data/service#data.api-key"
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldBeNil)
		So(values[id].Value, ShouldNotBeNil)
		So(string(values[id].Value), ShouldEqual, "service-api-key")
	})
}
