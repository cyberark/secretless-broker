package main

import (
	"testing"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/plugin/v1/testutils"
	"github.com/cyberark/secretless-broker/internal/providers"
)

func TestVault_Provider(t *testing.T) {
	var err error
	var provider plugin_v1.Provider
	name := "vault"

	options := plugin_v1.ProviderOptions{
		Name: name,
	}

	t.Run("Can create the Vault provider", func(t *testing.T) {
		provider, err = providers.ProviderFactories[name](options)
		assert.NoError(t, err)
	})

	t.Run("Has the expected provider name", func(t *testing.T) {
		assert.Equal(t, "vault", provider.GetName())
	})

	t.Run("Reports", func(t *testing.T) {
		for _, testCase := range reportsTestCases {
			t.Run(
				testCase.Description,
				testutils.Reports(provider, testCase.ID, testCase.ExpectedErrString),
			)
		}
	})

	t.Run("Provides", func(t *testing.T) {
		for _, testCase := range canProvideTestCases {
			t.Run(
				testCase.Description,
				testutils.CanProvide(provider, testCase.ID, testCase.ExpectedValue),
			)
		}
	})

	t.Run(
		"Multiple Provides ",
		testutils.CanProvideMultiple(
			provider,
			map[string]string{
				"cubbyhole/first-secret#some-key": "one",
				"cubbyhole/second-secret":         "two",
				"kv/db/password#password":         "two",
			},
		),
	)
}

var reportsTestCases = []testutils.ReportsTestCase{
	{
		Description: "Reports when the secret is not found",
		ID:          "foobar",
		ExpectedErrString: "HashiCorp Vault provider could not find secret " +
			"'foobar'",
	},
	{
		Description: "Reports when a field in the secret is not found",
		ID:          "cubbyhole/first-secret#foo.bar",
		ExpectedErrString: "HashiCorp Vault provider expects secret in " +
			"'foo.bar' at 'cubbyhole/first-secret'",
	},
}

var canProvideTestCases = []testutils.CanProvideTestCase{
	{
		Description:   "Can provide a cubbyhole secret",
		ID:            "cubbyhole/first-secret#some-key",
		ExpectedValue: "one",
	},
	{
		Description:   "Can provide a cubbyhole secret with default field name",
		ID:            "cubbyhole/second-secret",
		ExpectedValue: "two",
	},
	{
		Description:   "Can provide a KV v1 secret",
		ID:            "kv/db/password#password",
		ExpectedValue: "db-secret",
	},
	{
		Description:   "Can provide a KV v1 secret with default field name",
		ID:            "kv/web/password",
		ExpectedValue: "web-secret",
	},
	{
		// note the "data" in path and in the fields to navigate, which is required in KV v2
		Description:   "Can provide latest KV v2 secret",
		ID:            "secret/data/service#data.api-key",
		ExpectedValue: "service-api-key",
	},
}
