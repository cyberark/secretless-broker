package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/conjurinc/secretless/internal/pkg/provider"
	_ "github.com/joho/godotenv/autoload"
)

func TestProvider(t *testing.T) {
	var err error

	name := "vault"

	provider, err := provider.NewVaultProvider(name)
	if err != nil {
		t.Fatal(err)
	}

	Convey("Has the expected provider name", t, func() {
		So(provider.Name(), ShouldEqual, "vault")
	})

	Convey("Reports when the secret is not found", t, func() {
		value, err := provider.Value("foobar")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "HashiCorp Vault provider could not find a secret called 'foobar'")
		So(value, ShouldBeNil)
	})

	Convey("Can provide a secret", t, func() {
		value, err := provider.Value("kv/db/password")
		So(err, ShouldBeNil)
		So(string(value), ShouldEqual, "secret")
	})
}
