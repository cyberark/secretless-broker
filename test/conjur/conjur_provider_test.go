package main

import (
	"log"
	"io/ioutil"
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/cyberark/secretless-broker/pkg/secretless/config"
	"github.com/cyberark/secretless-broker/internal/pkg/plugin"
	"github.com/cyberark/secretless-broker/internal/app/secretless/providers"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"

	_ "github.com/joho/godotenv/autoload"
)

// TestConjur_Provider tests the ability of the ConjurProvider to provide a Conjur accessToken
// as well as secret values.
func TestConjur_Provider(t *testing.T) {
	var err error
	var provider plugin_v1.Provider
	name := "conjur"
	resolver := plugin.NewResolver(providers.ProviderFactories, nil, nil)

	Convey("Can be retrieved from the Resolver", t, func() {
		provider, err = resolver.GetProvider(name)
		So(err, ShouldBeNil)
		So(provider.GetName(), ShouldEqual, "conjur")
	})

	Convey("Can provide an access token", t, func() {
		value, err := provider.GetValue("accessToken")
		So(err, ShouldBeNil)

		token := make(map[string]string)
		err = json.Unmarshal(value, &token)
		So(err, ShouldBeNil)
		So(token["protected"], ShouldNotBeNil)
		So(token["payload"], ShouldNotBeNil)
	})

	Convey("Can provide a secret to a fully qualified variable", t, func() {
		value, err := provider.GetValue("dev:variable:db/password")
		So(err, ShouldBeNil)

		So(string(value), ShouldEqual, "secret")
	})

	Convey("Can provide the default Conjur account name", t, func() {
		value, err := provider.GetValue("variable:db/password")
		So(err, ShouldBeNil)

		So(string(value), ShouldEqual, "secret")
	})

	Convey("Can provide the default Conjur account name and resource type", t, func() {
		value, err := provider.GetValue("db/password")
		So(err, ShouldBeNil)

		So(string(value), ShouldEqual, "secret")
	})

	Convey("Cannot provide an unknown value", t, func() {
		_, err = provider.GetValue("foobar")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "404 Not Found. Variable 'foobar' not found in account 'dev'.")
	})

	Convey("Can provide multiple secrets", t, func() {
		values, err := resolver.Resolve([]config.Variable{
			config.Variable{
				Name: "user",
				Provider: name,
				ID: "db/user",
			},
			config.Variable{
				Name: "password",
				Provider: name,
				ID: "db/password", 
			},
			config.Variable{
				Name: "address",
				Provider: name,
				ID: "db/address", 
			},
		})

		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 3)
		So(string(values["user"]), ShouldEqual, "somewhat-secret")
		So(string(values["password"]), ShouldEqual, "secret")
		So(string(values["address"]), ShouldEqual, "not-so-secret")
	})

	Convey("Can provide multiple secrets including an access token", t, func() {
		values, err := resolver.Resolve([]config.Variable{
			config.Variable{
				Name: "user",
				Provider: name,
				ID: "db/user",
			},
			config.Variable{
				Name: "password",
				Provider: name,
				ID: "db/password", 
			},
			config.Variable{
				Name: "address",
				Provider: name,
				ID: "db/address", 
			},
			config.Variable{
				Name: "accessToken",
				Provider: name,
				ID: "accessToken", 
			},
		})
		
		So(err, ShouldBeNil)
		So(len(values), ShouldEqual, 4)
		So(string(values["accessToken"]), ShouldNotBeNil)
	})

	Convey("Can retrieve a single access token", t, func() {
		token, err := provider.GetValue("accessToken")
		So(err, ShouldBeNil)
		So(token, ShouldNotBeNil)
	})
}

var result map[string][]byte
func BenchmarkConjurProvider(b *testing.B) {
	b.StopTimer()

	log.SetOutput(ioutil.Discard)
	name := "conjur"
	resolver := plugin.NewResolver(providers.ProviderFactories, nil, nil)
	resolver.GetProvider(name)
	
	var values map[string][]byte

	b.StartTimer()

	for n := 0; n < b.N; n++ {
		values, _ = resolver.Resolve([]config.Variable{
			config.Variable{
				Name: "user",
				Provider: name,
				ID: "db/user",
			},
			config.Variable{
				Name: "password",
				Provider: name,
				ID: "db/password", 
			},
			config.Variable{
				Name: "address",
				Provider: name,
				ID: "db/address", 
			},
		})
	}
	result = values
}