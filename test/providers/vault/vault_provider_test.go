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

	Convey("Reports", t, func() {
		for _, testCase := range reportsTestCases {
			Convey(
				testCase.description,
				reports(provider, testCase.id, testCase.expectedErrString),
			)
		}
	})

	Convey("Provides", t, func() {
		for _, testCase := range canProvideTestCases {
			Convey(
				testCase.description,
				canProvide(provider, testCase.id, testCase.expectedValue),
			)
		}
	})
}

type canProvideTestCase struct {
	description   string
	id            string
	expectedValue string
}

func canProvide(provider plugin_v1.Provider, id string, expectedValue string) func() {
	return func() {
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldBeNil)
		So(values[id].Value, ShouldNotBeNil)
		So(string(values[id].Value), ShouldEqual, expectedValue)
	}
}

type reportsTestCase struct {
	description       string
	id                string
	expectedErrString string
}

func reports(provider plugin_v1.Provider, id string, expectedErrString string) func() {
	return func() {
		values, err := provider.GetValues(id)

		So(err, ShouldBeNil)
		So(values[id], ShouldNotBeNil)
		So(values[id].Error, ShouldNotBeNil)
		So(values[id].Error.Error(), ShouldEqual, expectedErrString)
		So(values[id].Value, ShouldBeNil)
	}
}

var reportsTestCases = []reportsTestCase{
	{
		description: "Reports when the secret is not found",
		id:          "foobar",
		expectedErrString: "HashiCorp Vault provider could not find secret " +
			"'foobar'",
	},
	{
		description: "Reports when a field in the secret is not found",
		id:          "cubbyhole/first-secret#foo.bar",
		expectedErrString: "HashiCorp Vault provider expects secret in " +
			"'foo.bar' at 'cubbyhole/first-secret'",
	},
}

var canProvideTestCases = []canProvideTestCase{
	{
		description:   "Can provide a cubbyhole secret",
		id:            "cubbyhole/first-secret#some-key",
		expectedValue: "one",
	},
	{
		description:   "Can provide a cubbyhole secret with default field name",
		id:            "cubbyhole/second-secret",
		expectedValue: "two",
	},
	{
		description:   "Can provide a KV v1 secret",
		id:            "kv/db/password#password",
		expectedValue: "db-secret",
	},
	{
		description:   "Can provide a KV v1 secret with default field name",
		id:            "kv/web/password",
		expectedValue: "web-secret",
	},
	{
		// note the "data" in path and in the fields to navigate, which is required in KV v2
		description:   "Can provide latest KV v2 secret",
		id:            "secret/data/service#data.api-key",
		expectedValue: "service-api-key",
	},
}
