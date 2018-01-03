package main

import (
  "encoding/json"
  "os"
  "testing"

  . "github.com/smartystreets/goconvey/convey"

  "github.com/kgilpin/secretless/internal/pkg/provider"
)


func TestProvider(t *testing.T) {
  var err error

  name := "conjur"

  configuration := make(map[string]string)
  configuration["url"] = "http://conjur"
  configuration["account"] = "dev"

  credentials := make(map[string]string)
  credentials["username"] = "admin"
  credentials["apiKey"] = os.Getenv("TEST_CONJUR_AUTHN_API_KEY")

  provider, err := provider.NewConjurProvider(name, configuration, credentials)
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

  Convey("Cannot provide an unknown value", t, func() {
    _, err = provider.Value("foobar")
    So(err, ShouldNotBeNil)
    So(err.Error(), ShouldEqual, "conjur does not know how to provide a value for 'foobar'")
  })

  Convey("Cannot provide an unloaded secret", t, func() {
    _, err = provider.Value("variable:foobar")
    So(err, ShouldNotBeNil)
    So(err.Error(), ShouldEqual, "404: Variable 'foobar' not found")
  })

  Convey("Can provide a secret", t, func() {
    value, err := provider.Value("variable:db/password")
    So(err, ShouldBeNil)

    So(string(value), ShouldEqual, "secret")
  })
}
