package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/cyberark/summon/secretsyml"
	. "github.com/smartystreets/goconvey/convey"

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

	Convey("Provides secrets to a subprocess environment", t, func() {
		provider := makePasswordProvider()
		subcommand := command.Subcommand{Args: []string{"env"}, Provider: provider, SecretsMap: makeDBPasswordSecretsMap()}
		stdout = captureStdoutFromSubcommand(&subcommand)

		err = subcommand.Run()
		lines := strings.Split(string(stdout.Bytes()), "\n")

		So(err, ShouldBeNil)
		So(lines, ShouldContain, "DB_PASSWORD=secret")
	})

	Convey("Echos a literal (non-secret) value", t, func() {
		secretsMap := make(map[string]secretsyml.SecretSpec)
		spec := secretsyml.SecretSpec{Path: "literal-secret", Tags: []secretsyml.YamlTag{secretsyml.Literal}}
		secretsMap["DB_PASSWORD"] = spec

		provider := makeEmptyProvider()
		subcommand := command.Subcommand{Args: []string{"env"}, Provider: provider, SecretsMap: secretsMap}
		stdout = captureStdoutFromSubcommand(&subcommand)

		err = subcommand.Run()
		lines := strings.Split(string(stdout.Bytes()), "\n")

		So(err, ShouldBeNil)
		So(lines, ShouldContain, "DB_PASSWORD=literal-secret")
	})

	Convey("Reports an error when the secrets cannot be found", t, func() {
		provider := makeEmptyProvider()
		subcommand := command.Subcommand{Args: []string{"env"}, Provider: provider, SecretsMap: makeDBPasswordSecretsMap()}

		err = subcommand.Run()

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Value 'db/password' not found in MapProvider")
	})

	Convey("Reports an error when the subprocess command is invalid", t, func() {
		provider := makeEmptyProvider()
		subcommand := command.Subcommand{Args: []string{"foobar"}, Provider: provider, SecretsMap: makeEmptySecretsMap()}

		err = subcommand.Run()

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, `exec: "foobar": executable file not found in $PATH`)
	})
}
