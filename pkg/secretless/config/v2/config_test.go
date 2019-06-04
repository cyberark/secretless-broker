package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const sampleConfigStr = `
version: 2
services:
  postgres-db:
    protocol: pg
    listenOn: tcp://0.0.0.0:5432 # can be a socket as well (same name for both)
    credentials:
      address: postgres.my-service.internal:5432
      password:
        from: vault
        get: name-in-vault
      username:
        from: env
        get: USERNAME
    config:  # this section usually blank
      optionalStuff: blah
`

func sampleConfig() (*Config, error) {
	configFileContents := []byte(sampleConfigStr)
	return NewConfig(configFileContents)
}

func TestNewConfig(t *testing.T) {
	t.Run("invalid file contents", func(t *testing.T) {
		configFileContents := []byte("12323232")
		_, err := NewConfig(configFileContents)
		assert.Error(t, err)
	})

	t.Run("blank file contents", func(t *testing.T) {
		configFileContents := []byte("")
		_, err := NewConfig(configFileContents)
		assert.Error(t, err)
	})

	t.Run("basic hydration", func(t *testing.T) {
		cfg, err := sampleConfig()
		assert.NoError(t, err)
		if err != nil {
			return
		}

		assert.Equal(t, "postgres-db", cfg.Services[0].Name)
		assert.Equal(t, "pg", cfg.Services[0].Protocol)
		assert.Equal(t, "tcp://0.0.0.0:5432", cfg.Services[0].ListenOn)
	})

	t.Run("config hydration", func(t *testing.T) {
		cfg, err := sampleConfig()
		assert.NoError(t, err)
		if err != nil {
			return
		}

		expectedBytes := []byte("optionalStuff: blah\n")
		assert.Equal(t, expectedBytes, cfg.Services[0].ProtocolConfig)
	})

	t.Run("credential hydration", func(t *testing.T) {
		cfg, err := sampleConfig()
		assert.NoError(t, err)
		if err != nil {
			return
		}

		actualCreds := cfg.Services[0].Credentials
		expectedCreds := []*Credential{
			{
				Name: "address",
				From: "literal",
				Get:  "postgres.my-service.internal:5432",
			},
			{
				Name: "password",
				From: "vault",
				Get:  "name-in-vault",
			},
			{
				Name: "username",
				From: "env",
				Get:  "USERNAME",
			},
		}
		assert.ElementsMatch(t, expectedCreds, actualCreds)
	})
}
