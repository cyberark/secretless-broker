package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	yaml "gopkg.in/yaml.v1"

	"github.com/conjurinc/secretless/internal/pkg/provider"
)

func TestProvider(t *testing.T) {
	var err error

	// Utility to load the config that is stored by test.sh
	// This provides a means for running a native Go environment with
	// Vault running in a container.
	type VaultConfig struct {
		Address string
		Token   string
	}

	vaultrcFile := "./tmp/.vaultrc"
	name := "vault"

	_, err = os.Stat(vaultrcFile)
	if os.IsNotExist(err) {
		panic(fmt.Sprintf("vaultrc file %s does not exist; run ./start.sh to create it", vaultrcFile))
	}

	vaultConfig := VaultConfig{}
	buf, err := ioutil.ReadFile(vaultrcFile)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(buf, &vaultConfig)
	if err != nil {
		panic(err)
	}

	url := os.Getenv("VAULT_ADDR")

	configuration := make(map[string]string)
	if url != "" {
		configuration["address"] = url
	} else {
		configuration["address"] = vaultConfig.Address
	}

	credentials := make(map[string]string)
	credentials["token"] = vaultConfig.Token

	provider, err := provider.NewVaultProvider(name, configuration, credentials)
	if err != nil {
		t.Fatal(err)
	}

	Convey("Has the expected provider name", t, func() {
		So(provider.Name(), ShouldEqual, "vault")
	})

	Convey("Reports when the secret is not found", t, func() {
		value, err := provider.Value("foobar")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "HashiCorp Vault provider could not find a secret called 'foobar'")
		So(value, ShouldBeNil)
	})

	Convey("Can provide a secret", t, func() {
		value, err := provider.Value("kv/db/password")
		So(err, ShouldBeNil)
		So(string(value), ShouldEqual, "secret")
	})
}
