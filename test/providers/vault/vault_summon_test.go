package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"

	"github.com/cyberark/secretless-broker/internal/summon/command"
)

func TestVault_Summon(t *testing.T) {
	secretsDescriptor := `
FIRST_SECRET: !var cubbyhole/first-secret#some-key
SECOND_SECRET: !var cubbyhole/second-secret
DB_PASSWORD: !var kv/db/password#password
WEB_PASSWORD: !var kv/web/password
SVC_API_KEY: !var secret/data/service#data.api-key
`
	defaultArgs := []string{"summon2", "-p", "vault", "--yaml", secretsDescriptor, "env"}

	runCommand := func(args []string) (lines []string, err error) {
		var buffer bytes.Buffer
		writer := bufio.NewWriter(&buffer)

		err = command.RunCLI(args, writer)

		writer.Flush()
		output := string(buffer.Bytes())
		lines = strings.Split(output, "\n")

		return
	}

	t.Run("Can summon secrets from Vault", func(t *testing.T) {
		lines, err := runCommand(defaultArgs)

		assert.NoError(t, err)
		assert.Contains(t, lines, "FIRST_SECRET=one")
		assert.Contains(t, lines, "SECOND_SECRET=two")
		assert.Contains(t, lines, "DB_PASSWORD=db-secret")
		assert.Contains(t, lines, "WEB_PASSWORD=web-secret")
		assert.Contains(t, lines, "SVC_API_KEY=service-api-key")
	})
}
