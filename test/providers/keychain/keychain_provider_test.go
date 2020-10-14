package main

import (
	"os"
	"testing"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/plugin/v1/testutils"
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

	// e.g. ${service}_1#${account}_1
	getSecretPath := func(idx int) string {
		return service + "_" + string(idx) + "#" + account + "_" + string(idx)
	}
	// e.g. ${secret}_1
	getSecretValue := func(idx int) string {
		return secret + "_" + string(idx)
	}

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

	Convey(
		"Can provide a valid secret value",
		t,
		testutils.CanProvide(
			provider,
			getSecretPath(1),
			getSecretValue(1),
		),
	)

	Convey(
		"Multiple Provides ",
		t,
		testutils.CanProvideMultiple(
			provider,
			map[string]string{
				getSecretPath(1): getSecretValue(1),
				getSecretPath(2): getSecretValue(2),
				getSecretPath(3): getSecretValue(3),
			},
		),
	)

	Convey(
		"Returns an error for an invalid secret value",
		t,
		testutils.CanProvide(
			provider,
			"madeup#secret",
			"The specified item could not be found in the keychain.",
		),
	)
}
