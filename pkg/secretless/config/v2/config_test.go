package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const sampleConfigWithProtocolStr = `
version: 2
services:
  postgres-db:
    protocol: pg
    listenOn: tcp://0.0.0.0:5432 # can be a socket as well (same name for both)
    credentials:
      host: postgres.my-service.internal
      password:
        from: vault
        get: name-in-vault
      username:
        from: env
        get: USERNAME
    config:  # this section usually blank
      optionalStuff: blah
  aws-proxy:
    protocol: http
    listenOn: tcp://0.0.0.0:8080
    credentials:
      accessKeyId:
        from: env
        get: AWS_ACCESS_KEY_ID
      secretAccessKey:
        from: env
        get: AWS_SECRET_ACCESS_KEY
    config:
      authenticationStrategy: aws
      authenticateURLsMatching:
        - .*
`

const sampleConfigWithConnectorStr = `
version: 2
services:
  postgres-db:
    connector: pg
    listenOn: tcp://0.0.0.0:5432 # can be a socket as well (same name for both)
    credentials:
      host: postgres.my-service.internal
      password:
        from: vault
        get: name-in-vault
      username:
        from: env
        get: USERNAME
    config:  # this section usually blank
      optionalStuff: blah
  aws-proxy:
    connector: aws
    listenOn: tcp://0.0.0.0:8080
    credentials:
      accessKeyId:
        from: env
        get: AWS_ACCESS_KEY_ID
      secretAccessKey:
        from: env
        get: AWS_SECRET_ACCESS_KEY
    config:
      authenticateURLsMatching:
        - .*
`

func sampleConfig(contents string) (*Config, error) {
	configFileContents := []byte(contents)
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

	RunNewConfigTestCases(t, "with-protocol", sampleConfigWithProtocolStr)
	RunNewConfigTestCases(t, "with-connector", sampleConfigWithConnectorStr)
}

func RunNewConfigTestCases(t *testing.T, label string, sampleContents string) {
	t.Run(label+": basic hydration", func(t *testing.T) {
		cfg, err := sampleConfig(sampleContents)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		assert.Equal(t, "postgres-db", cfg.Services[1].Name)
		assert.Equal(t, "pg", cfg.Services[1].Connector)
		assert.Equal(t, NetworkAddress("tcp://0.0.0.0:5432"), cfg.Services[1].ListenOn)
	})

	t.Run(label+": config hydration", func(t *testing.T) {
		cfg, err := sampleConfig(sampleContents)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		expectedBytes := []byte("optionalStuff: blah\n")
		assert.Equal(t, string(expectedBytes), string(cfg.Services[1].ConnectorConfig))
	})

	t.Run(label+": http hydration", func(t *testing.T) {
		cfg, err := sampleConfig(sampleContents)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		assert.Equal(t, "aws-proxy", cfg.Services[0].Name)
		assert.Equal(t, "aws", cfg.Services[0].Connector)
		assert.Equal(t, NetworkAddress("tcp://0.0.0.0:8080"), cfg.Services[0].ListenOn)
	})

	t.Run(label+": credential hydration", func(t *testing.T) {
		cfg, err := sampleConfig(sampleContents)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		actualCreds := cfg.Services[1].Credentials
		expectedCreds := []*Credential{
			{
				Name: "host",
				From: "literal",
				Get:  "postgres.my-service.internal",
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
