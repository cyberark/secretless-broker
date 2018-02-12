package main

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/conjurinc/secretless/internal/pkg/provider"
	_ "github.com/joho/godotenv/autoload"
)

// TestConjur_Provider tests the ability of the ConjurProvider to provide a Conjur accessToken
// as well as secret values.
func TestConjur_Provider(t *testing.T) {
	name := "conjur"

	provider, err := provider.NewConjurProvider(name)
	if err != nil {
		t.Fatal(err)
	}

	Convey("Has the expected provider name", t, func() {
		So(provider.Name(), ShouldEqual, "conjur")
	})

	Convey("Can provide an access token", t, func() {
		value, err := provider.Value("accessToken")
		So(err, ShouldBeNil)

		token := make(map[string]string)
		err = json.Unmarshal(value, &token)
		So(err, ShouldBeNil)
		So(token["protected"], ShouldNotBeNil)
		So(token["payload"], ShouldNotBeNil)
	})

	Convey("Can provide a secret", t, func() {
		value, err := provider.Value("variable:db/password")
		So(err, ShouldBeNil)

		So(string(value), ShouldEqual, "secret")
	})

	Convey("Cannot provide an unknown value", t, func() {
		_, err = provider.Value("foobar")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "conjur does not know how to provide a value for 'foobar'")
	})

	Convey("Cannot provide an unloaded secret", t, func() {
		_, err = provider.Value("variable:foobar")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Variable 'foobar' not found in account 'dev'")
	})
}
