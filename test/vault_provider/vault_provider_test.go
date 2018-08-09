package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/cyberark/secretless-broker/internal/app/secretless/providers"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"

	_ "github.com/joho/godotenv/autoload"
)

func TestVault_Provider(t *testing.T) {
	name := "vault"

	options := plugin_v1.ProviderOptions{
		Name: name,
	}

	provider := providers.ProviderFactories[name](options)

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
