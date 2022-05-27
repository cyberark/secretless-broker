package main

import (
	"encoding/json"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/plugin/v1/testutils"
	"github.com/cyberark/secretless-broker/internal/providers"
)

// TestConjur_Provider tests the ability of the ConjurProvider to provide a Conjur accessToken
// as well as secret values.
func TestConjur_Provider(t *testing.T) {
	var err error
	var provider plugin_v1.Provider
	name := "conjur"

	options := plugin_v1.ProviderOptions{
		Name: name,
	}

	t.Run("Can create the Conjur provider", func(t *testing.T) {
		provider, err = providers.ProviderFactories[name](options)
		assert.NoError(t, err)
	})

	t.Run("Has the expected provider name", func(t *testing.T) {
		assert.Equal(t, "conjur", provider.GetName())
	})

	t.Run("Can provide an access token", func(t *testing.T) {
		id := "accessToken"
		values, err := provider.GetValues(id)

		assert.NoError(t, err)
		assert.NotNil(t, values[id])
		assert.NoError(t, values[id].Error)
		assert.NotNil(t, values[id].Value)

		token := make(map[string]string)
		err = json.Unmarshal(values[id].Value, &token)
		assert.NoError(t, err)
		assert.NotNil(t, token["protected"])
		assert.NotNil(t, token["payload"])
	})

	t.Run("Reports an unknown value",
		testutils.Reports(
			provider,
			"foobar",
			"404 Not Found. CONJ00076E Variable dev:variable:foobar is empty or not found..",
		),
	)

	t.Run("Provides", func(t *testing.T) {
		for _, testCase := range canProvideTestCases {
			t.Run(testCase.Description, testutils.CanProvide(provider, testCase.ID, testCase.ExpectedValue))
		}
	})
}

var canProvideTestCases = []testutils.CanProvideTestCase{
	{
		Description:   "Can provide a secret to a fully qualified variable",
		ID:            "dev:variable:db/password",
		ExpectedValue: "secret",
	},
	{
		Description:   "Can retrieve a secret value with spaces",
		ID:            "my var",
		ExpectedValue: "othersecret",
	},
	{
		Description:   "Can provide the default Conjur account name",
		ID:            "variable:db/password",
		ExpectedValue: "secret",
	},
	{
		Description:   "Can provide the default Conjur account name and resource type",
		ID:            "db/password",
		ExpectedValue: "secret",
	},
}
