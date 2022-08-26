package main

import (
	"testing"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/plugin/v1/testutils"
	"github.com/cyberark/secretless-broker/internal/providers"
	"github.com/stretchr/testify/assert"
)

func TestKeychainProvider(t *testing.T) {
	// Setup.

	// Create all the keychain items here.
	//
	// It's necessary to do this here because the keychain automatically trusts the
	// process that writes the secret. Without this a user would need confirm a keychain
	// prompt at least once before a read is possible.
	cleanup()
	defer cleanup()
	if err := setup(); err != nil {
		t.Fatal(err)
	}

	// Testing.

	providerName := "keychain"

	provider, err := providers.ProviderFactories[providerName](plugin_v1.ProviderOptions{
		Name: providerName,
	})
	if err != nil {
		// there was an error creating the provider, so exit the tests
		t.Error("Unable to create keychain provider.")
		t.FailNow()
	}

	t.Run("Has the expected provider name", func(t *testing.T) {
		assert.Equal(t, providerName, provider.GetName())
	})

	t.Run(
		"Can provide a valid secret value",
		testutils.CanProvide(
			provider,
			getSecretPath(1),
			getSecretValue(1),
		),
	)

	t.Run(
		"Multiple Provides ",
		testutils.CanProvideMultiple(
			provider,
			map[string]string{
				getSecretPath(1): getSecretValue(1),
				getSecretPath(2): getSecretValue(2),
				getSecretPath(3): getSecretValue(3),
			},
		),
	)

	t.Run(
		"Returns an error for an invalid secret value",
		testutils.Reports(
			provider,
			"madeup#secret",
			"The specified item could not be found in the keychain.",
		),
	)
}
