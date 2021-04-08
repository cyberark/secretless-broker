package main

import (
	"encoding/json"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	. "github.com/smartystreets/goconvey/convey"

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

	Convey("Can create the Conjur provider", t, func() {
		provider, err = providers.ProviderFactories[name](options)
		So(err, ShouldBeNil)
	})

	Convey("Has the expected provider name", t, func() {
		So(provider.GetName(), ShouldEqual, "conjur")
	})

	Convey("Can provide an access token", t, func() {
		id := "accessToken"
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldBeNil)
		So(values[id].Value, ShouldNotBeNil)

		token := make(map[string]string)
		err = json.Unmarshal(values[id].Value, &token)
		So(err, ShouldBeNil)
		So(token["protected"], ShouldNotBeNil)
		So(token["payload"], ShouldNotBeNil)
	})

	Convey(
		"Reports an unknown value",
		t,
		testutils.Reports(
			provider,
			"foobar",
			"404 Not Found. CONJ00076E Variable dev:variable:foobar is empty or not found..",
		),
	)

	Convey("Provides", t, func() {
		for _, testCase := range canProvideTestCases {
			Convey(
				testCase.Description,
				testutils.CanProvide(provider, testCase.ID, testCase.ExpectedValue),
			)
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
