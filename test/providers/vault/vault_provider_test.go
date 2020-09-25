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
		value, err := provider.GetValue("foobar")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "HashiCorp Vault provider could not find secret 'foobar'")
		So(value, ShouldBeNil)
	})

	Convey("Reports when a field in the secret is not found", t, func() {
		value, err := provider.GetValue("cubbyhole/first-secret#foo.bar")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "HashiCorp Vault provider expects secret in 'foo.bar' at 'cubbyhole/first-secret'")
		So(value, ShouldBeNil)
	})

	Convey("Can provide a cubbyhole secret", t, func() {
		value, err := provider.GetValue("cubbyhole/first-secret#some-key")
		So(err, ShouldBeNil)
		So(string(value), ShouldEqual, "one")
	})

	Convey("Can provide a cubbyhole secret with default field name", t, func() {
		value, err := provider.GetValue("cubbyhole/second-secret")
		So(err, ShouldBeNil)
		So(string(value), ShouldEqual, "two")
	})

	Convey("Can provide a KV v1 secret", t, func() {
		value, err := provider.GetValue("kv/db/password#password")
		So(err, ShouldBeNil)
		So(string(value), ShouldEqual, "db-secret")
	})

	Convey("Can provide a KV v1 secret with default field name", t, func() {
		value, err := provider.GetValue("kv/web/password")
		So(err, ShouldBeNil)
		So(string(value), ShouldEqual, "web-secret")
	})

	// note the "data" in path and in the fields to navigate, which is required in KV v2
	Convey("Can provide latest KV v2 secret", t, func() {
		value, err := provider.GetValue("secret/data/service#data.api-key")
		So(err, ShouldBeNil)
		So(string(value), ShouldEqual, "service-api-key")
	})
}
