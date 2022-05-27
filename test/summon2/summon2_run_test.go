package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/cyberark/summon/secretsyml"
	"github.com/stretchr/testify/assert"

	plugin_v1 "github.com/cyberark/secretless-broker/internal/plugin/v1"
	"github.com/cyberark/secretless-broker/internal/summon/command"
)

type MapProvider struct {
	Secrets map[string][]byte
}

func (mp MapProvider) GetName() string {
	return "mapProvider"
}

// GetValues takes in variable ids and returns their resolved values. This method is
// needed to the Provider interface
func (mp MapProvider) GetValues(ids ...string) (map[string]plugin_v1.ProviderResponse, error) {
	return plugin_v1.GetValues(mp, ids...)
}

func (mp MapProvider) GetValue(id string) ([]byte, error) {
	value, ok := mp.Secrets[id]
	if ok {
		return value, nil
	}
	return nil, fmt.Errorf("Value '%s' not found in MapProvider", id)
}

func makeEmptyProvider() plugin_v1.Provider {
	secrets := make(map[string][]byte)
	return MapProvider{Secrets: secrets}
}

func makePasswordProvider() plugin_v1.Provider {
	secrets := make(map[string][]byte)
	secrets["db/password"] = []byte("secret")
	return MapProvider{Secrets: secrets}
}

func makeDBPasswordSecretsMap() (secretsMap secretsyml.SecretsMap) {
	secretsMap = make(map[string]secretsyml.SecretSpec)
	spec := secretsyml.SecretSpec{Path: "db/password", Tags: []secretsyml.YamlTag{secretsyml.Var}}
	secretsMap["DB_PASSWORD"] = spec
	return
}

func makeEmptySecretsMap() (secretsMap secretsyml.SecretsMap) {
	secretsMap = make(map[string]secretsyml.SecretSpec)
	return
}

func captureStdoutFromSubcommand(sc *command.Subcommand) *bytes.Buffer {
	var b bytes.Buffer
	sc.Stdout = bufio.NewWriter(&b)
	return &b
}

// TestSummon2_Run tests the Command.Run capability. This is a lower level than the CLI.
func TestSummon2_Run(t *testing.T) {
	var stdout *bytes.Buffer
	var err error

	t.Run("Provides secrets to a subprocess environment", func(t *testing.T) {
		provider := makePasswordProvider()
		subcommand := command.Subcommand{Args: []string{"env"}, Provider: provider, SecretsMap: makeDBPasswordSecretsMap()}
		stdout = captureStdoutFromSubcommand(&subcommand)

		err = subcommand.Run()
		lines := strings.Split(string(stdout.Bytes()), "\n")

		assert.NoError(t, err)
		assert.Contains(t, lines, "DB_PASSWORD=secret")
	})

	t.Run("Echos a literal (non-secret) value", func(t *testing.T) {
		secretsMap := make(map[string]secretsyml.SecretSpec)
		spec := secretsyml.SecretSpec{Path: "literal-secret", Tags: []secretsyml.YamlTag{secretsyml.Literal}}
		secretsMap["DB_PASSWORD"] = spec

		provider := makeEmptyProvider()
		subcommand := command.Subcommand{Args: []string{"env"}, Provider: provider, SecretsMap: secretsMap}
		stdout = captureStdoutFromSubcommand(&subcommand)

		err = subcommand.Run()
		lines := strings.Split(string(stdout.Bytes()), "\n")

		assert.NoError(t, err)
		assert.Contains(t, lines, "DB_PASSWORD=literal-secret")
	})

	testCases := []struct {
		name          string
		args          []string
		secretsMap    secretsyml.SecretsMap
		expectedError string
	}{
		{
			name:          "Reports an error when the secrets cannot be found",
			args:          []string{"env"},
			secretsMap:    makeDBPasswordSecretsMap(),
			expectedError: "Value 'db/password' not found in MapProvider",
		},
		{
			name:          "Reports an error when the subprocess command is invalid",
			args:          []string{"foobar"},
			secretsMap:    makeEmptySecretsMap(),
			expectedError: "exec: \"foobar\": executable file not found in $PATH",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := makeEmptyProvider()
			subcommand := command.Subcommand{Args: tc.args, Provider: provider, SecretsMap: tc.secretsMap}

			err = subcommand.Run()

			assert.Error(t, err)
			assert.EqualError(t, err, tc.expectedError)
		})
	}
}
