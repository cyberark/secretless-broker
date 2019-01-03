package main

import (
	"log"

	keychainProvider "github.com/cyberark/secretless-broker/internal/app/secretless/providers/keychain"
	plugin_v1 "github.com/cyberark/secretless-broker/pkg/secretless/plugin/v1"
)

func main() {
	options := plugin_v1.ProviderOptions{
		Name: "foo",
	}

	provider, provierErr := keychainProvider.ProviderFactory(options)
	if provierErr != nil {
		panic(provierErr)
	}
	keyValue, keyErr := provider.GetValue("foo.bar.item#accountname")
	if keyErr != nil {
		panic(keyErr)
	}

	log.Printf("Bytes: %v\n", keyValue)
	log.Printf("String: %s\n", keyValue)
}
