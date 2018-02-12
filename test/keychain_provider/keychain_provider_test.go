package main

import (
	"os/exec"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/conjurinc/secretless/internal/app/secretless/variable"
	"github.com/conjurinc/secretless/internal/pkg/provider/keychain_provider"
	"github.com/conjurinc/secretless/pkg/secretless/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestKeychainProvider(t *testing.T) {

	Convey("With a secret stored in the keychain", t, func() {
		service := "Secretless_TestKeychainProvider"
		account := "password"
		secret := "secret"

		// $ security delete-generic-password -a password -s Secretless_TestKeychainProvider

		// $ security add-generic-password -a password -s Secretless_TestKeychainProvider -w secret

		// $ security find-generic-password -a password -s Secretless_TestKeychainProvider -w
		// secret

		exec.Command("security", "delete-generic-password", "-a", account, "-s", service).CombinedOutput()

		output, err := exec.Command("security", "add-generic-password", "-a", account, "-s", service, "-w", secret).CombinedOutput()
		So(string(output), ShouldEqual, "")
		So(err, ShouldBeNil)

		Convey("The secret value can be obtained directly as GetGenericPassword", func() {
			obtainedPassword, err := keychain_provider.GetGenericPassword(service, account)
			So(err, ShouldBeNil)
			So(string(obtainedPassword), ShouldEqual, secret)
		})

		Convey("The secret value can be obtaind through the provider interface", func() {
			id := strings.Join([]string{service, account}, "#")
			v := config.Variable{ID: id, Provider: "keychain", Name: "password"}

			values, err := variable.Resolve([]config.Variable{v})
			So(err, ShouldBeNil)

			expected := make(map[string]string)
			expected["password"] = secret

			actual := make(map[string]string)
			for k, v := range values {
				actual[k] = string(v)
			}

			actual_s, _ := yaml.Marshal(actual)
			expected_s, _ := yaml.Marshal(expected)

			So(string(actual_s), ShouldEqual, string(expected_s))
		})
	})
}
