package main

import (
	"strconv"

	"github.com/keybase/go-keychain"
)

const service = "TestGenericPasswordRef"
const account = "test"
const secret = "toomanysecrets"

// e.g. ${service}_1#${account}_1
func getSecretPath(idx int) string {
	return getService(idx) + "#" + getAccount(idx)
}

// e.g. ${account}_1
func getAccount(idx int) string {
	return account + "_" + strconv.Itoa(idx)
}

// e.g. ${service}_1
func getService(idx int) string {
	return service + "_" + strconv.Itoa(idx)
}

// e.g. ${secret}_1
func getSecretValue(idx int) string {
	return secret + "_" + strconv.Itoa(idx)
}

const numTestSecrets = 3

// setup populates the keychain with test secrets
func setup() error {
	// Create all the keychain items here.
	//
	// It's necessary to do this inside the test process so that the keychain
	// automatically trusts the process that writes the secret. Without this a
	// user would need confirm a keychain prompt at least once before a read is possible.
	for idx := 1; idx <= numTestSecrets; idx++ {
		item := keychain.NewGenericPassword(
			getService(idx),
			getAccount(idx),
			"",
			[]byte(getSecretValue(idx)),
			"",
		)

		if err := keychain.AddItem(item); err != nil {
			return err
		}
	}

	return nil
}

// cleanup remove all the test secrets from the keychain
func cleanup() {
	for idx := 1; idx <= numTestSecrets; idx++ {
		item := keychain.NewGenericPassword(
			getService(idx),
			getAccount(idx),
			"",
			[]byte(getSecretValue(idx)),
			"",
		)
		_ = keychain.DeleteItem(item)
	}
}
