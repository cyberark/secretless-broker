package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// reflex -r '\.go$' -s -- bash -c "go test"

const sampleConfig = `
version: "1"
services:
  postgres-db:
    protocol: pg
    listenOn: 0.0.0.0:5432 # can be a socket as well (same name for both)
    credentials:
      address: postgres.my-service.internal:5432
      password:
        providerId: name-in-vault
        provider: vault
      username:
        providerId: username
        provider: env
    config:  # this section usually blank
      optionalStuff: blah
`


func TestNewConfig(t *testing.T) {

	t.Run("invalid file contents", func(t *testing.T) {
		configFileContents := []byte("12323232")
		_, err := NewConfigV2(configFileContents)
		assert.Error(t, err)
	})

	t.Run("blank file contents", func(t *testing.T) {
		configFileContents := []byte("")
		_, err := NewConfigV2(configFileContents)
		assert.Error(t, err)
	})

	t.Run("valid file contents", func(t *testing.T) {
		configFileContents := []byte(sampleConfig)
		_, err := NewConfigV2(configFileContents)
		assert.Nil(t, err)
		// TODO: add sanity check that hydration occurred?
	})

}

func TestConvertToV1(t *testing.T) {
	// standard case
	t.Run("valid file contents", func(t *testing.T) {
		configFileContents := []byte(sampleConfig)
		_, err := NewConfigV2(configFileContents)
		assert.Nil(t, err)
		// TODO: add sanity check that hydration occurred?
	})

	// liston on socket and tcp

	// provider literals

}
