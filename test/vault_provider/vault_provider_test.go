package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/providers"

	_ "github.com/joho/godotenv/autoload"
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
		So(err.Error(), ShouldEqual, "HashiCorp Vault provider could not find a secret called 'foobar'")
		So(value, ShouldBeNil)
	})

	Convey("Can provide a secret", t, func() {
		value, err := provider.GetValue("kv/db/password#password")
		So(err, ShouldBeNil)
		So(string(value), ShouldEqual, "db-secret")
	})

	Convey("Can provide a secret with default field name", t, func() {
		value, err := provider.GetValue("kv/web/password")
		So(err, ShouldBeNil)
		So(string(value), ShouldEqual, "web-secret")
	})
}
