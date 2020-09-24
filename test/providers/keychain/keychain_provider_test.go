package main

import (
	"os"
	"strings"
	"testing"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/providers"

	. "github.com/smartystreets/goconvey/convey"
)

func TestKeychainProvider(t *testing.T) {
	var err error
	var provider plugin_v1.Provider

	name := "keychain"

	// get the environment variables with the test config
	service := os.Getenv("SERVICE")
	account := os.Getenv("ACCOUNT")
	secret := os.Getenv("SECRET")

	options := plugin_v1.ProviderOptions{
		Name: name,
	}

	provider, err = providers.ProviderFactories[name](options)
	if err != nil {
		// there was an error creating the provider, so exit the tests
		t.Error("Unable to create keychain provider.")
		t.FailNow()
	}

	Convey("Has the expected provider name", t, func() {
		So(provider.GetName(), ShouldEqual, name)
	})

	Convey("Can provide a valid secret value", t, func() {
		id := strings.Join([]string{service, account}, "#")

		values, err := provider.GetValues(id)
		So(err, ShouldBeNil)
		So(string(values[0]), ShouldEqual, secret)
	})

	Convey("Returns an error for an invalid secret value", t, func() {
		id := "madeup#secret"

		_, err := provider.GetValues(id)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "The specified item could not be found in the keychain.")
	})
}
