package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	yaml "gopkg.in/yaml.v1"

	"github.com/kgilpin/secretless/internal/pkg/provider"
)

func TestProvider(t *testing.T) {
	var err error

	// Utility to load the config that is stored by test.sh
	// This provides a means for running a native Go environment with
	// Conjur running in a container.
	type ConjurConfig struct {
		URL     string
		Account string
		APIKey  string `yaml:"api_key"`
	}

	conjurrcFile := "./tmp/.conjurrc"
	name := "conjur"

	_, err = os.Stat(conjurrcFile)
	if os.IsNotExist(err) {
		panic(fmt.Sprintf("conjurrc file %s does not exist; run ./start.sh to create it", conjurrcFile))
	}

	conjurConfig := ConjurConfig{}
	buf, err := ioutil.ReadFile(conjurrcFile)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(buf, &conjurConfig)
	if err != nil {
		panic(err)
	}

	url := os.Getenv("CONJUR_APPLIANCE_URL")

	configuration := make(map[string]string)
	if url != "" {
		configuration["url"] = url
	} else {
		configuration["url"] = conjurConfig.URL
	}
	configuration["account"] = conjurConfig.Account

	credentials := make(map[string]string)
	credentials["username"] = "admin"
	credentials["apiKey"] = conjurConfig.APIKey

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
