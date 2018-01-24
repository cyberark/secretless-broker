package main

import (
	"os/exec"
	"testing"

	"github.com/conjurinc/secretless/internal/pkg/keychain"

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

		Convey("The secret value can be obtained", func() {
			obtainedPassword, err := keychain.GetGenericPassword(service, account)
			So(err, ShouldBeNil)
			So(obtainedPassword, ShouldEqual, secret)
		})
	})
}
