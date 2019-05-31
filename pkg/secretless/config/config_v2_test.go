package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
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

func sampleConfig() (*ConfigV2, error) {
	configFileContents := []byte(sampleConfigStr)
	return NewConfigV2(configFileContents)
}

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

	t.Run("basic hydration", func(t *testing.T) {
		cfg, err := sampleConfig()
		assert.NoError(t, err)

		assert.Equal(t, "postgres-db", cfg.Services[0].Name)
		assert.Equal(t, "pg", cfg.Services[0].Protocol)
		assert.Equal(t, "tcp://0.0.0.0:5432", cfg.Services[0].ListenOn)
	})

	t.Run("config hydration", func(t *testing.T) {
		cfg, err := sampleConfig()
		assert.NoError(t, err)

		expectedBytes := []byte("optionalStuff: blah\n")
		assert.Equal(t, expectedBytes, cfg.Services[0].Config)
	})

	t.Run("credential hydration", func(t *testing.T) {
		cfg, err := sampleConfig()
		assert.NoError(t, err)

		actualCreds := cfg.Services[0].Credentials
		expectedCreds := []*Credential{
			{
				Name: "address",
				From: "literal",
				Get: "postgres.my-service.internal:5432",
			},
			{
				Name: "password",
				From: "vault",
				Get: "name-in-vault",
			},
			{
				Name: "username",
				From: "env",
				Get: "USERNAME",
			},
		}
		assert.ElementsMatch(t, expectedCreds, actualCreds)
	})
}
